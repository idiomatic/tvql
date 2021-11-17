package model

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/idiomatic/tvql/metadata/mp4"
	"github.com/sunfish-shogi/bufseekio"
)

type Metavideo struct {
	Video
	Path string
}

type Metarendition struct {
	Path string
}

// useful for toll-free use of encoded VideoID/EpisodeID via client
type Stringer struct {
	S string
}

func (s Stringer) String() string {
	return s.S
}

type MetavideoID interface {
	fmt.Stringer
}

type VideoID struct {
	Title       string
	ReleaseYear int
}

// note: exposed to client
func (id VideoID) String() string {
	key := fmt.Sprintf("%s (%d)", id.Title, id.ReleaseYear)
	return key
	return hashToStr(key)
}

type SeriesID string

type SeasonID struct {
	SeriesID     SeriesID
	SeasonNumber int
}

type EpisodeID struct {
	SeasonID
	EpisodeNumber int
	VideoID
}

// note: exposed to client
func (id EpisodeID) String() string {
	key := fmt.Sprintf("%s %d%02d %s", id.SeriesID, id.SeasonNumber, id.EpisodeNumber, id.VideoID)
	return key
	return hashToStr(key)
}

type Library struct {
	Metavideos     map[MetavideoID]*Metavideo
	Series         map[SeriesID]*Series
	Seasons        map[SeasonID]*Season
	Metarenditions map[string]*Metarendition
	Mutex          sync.Mutex
}

func NewLibrary() *Library {
	return &Library{
		Metavideos:     make(map[MetavideoID]*Metavideo),
		Series:         make(map[SeriesID]*Series),
		Seasons:        make(map[SeasonID]*Season),
		Metarenditions: make(map[string]*Metarendition),
	}
}

func (l *Library) Survey(root string) error {
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ext := filepath.Ext(path)
			if ext != ".m4v" {
				return nil
			}

			return func() error {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				bufferedFile := bufseekio.NewReadSeeker(file, 1024, 4)

				videoFile := mp4.NewFile(bufferedFile)

				title, err := videoFile.Title()
				if err != nil || title == "" {
					// HACK
					title = strings.TrimSuffix(filepath.Base(path), ext)
				}

				releaseDate, err := videoFile.ReleaseDate()
				if err != nil || releaseDate == "" {
					// HACK
					releaseDate = "1900"
				}
				releaseYear, _ := strconv.Atoi(releaseDate[:4])

				seriesName, seriesNameErr := videoFile.TVShowName()
				seasonNumber, seasonNumberErr := videoFile.TVSeason()
				episodeNumber, episodeNumberErr := videoFile.TVEpisode()

				var (
					videoID   = VideoID{title, releaseYear}
					seriesID  = SeriesID(seriesName)
					seasonID  = SeasonID{seriesID, seasonNumber}
					episodeID = EpisodeID{seasonID, episodeNumber, videoID}

					metavideoID MetavideoID = videoID
				)

				if seriesNameErr == nil && seasonNumberErr == nil && episodeNumberErr == nil && seriesName != "" && seasonNumber > 0 {
					metavideoID = episodeID
				}

				l.Mutex.Lock()
				metavideo, ok := l.Metavideos[videoID]
				if !ok {
					metavideo = &Metavideo{
						Video{
							ID:          metavideoID.String(),
							Title:       title,
							ReleaseYear: releaseYear,
						},
						path,
					}
					l.Metavideos[metavideoID] = metavideo
				}
				l.Mutex.Unlock()

				video := &metavideo.Video

				if sortTitle, err := videoFile.SortTitle(); err != nil || sortTitle == "" {
					video.SortTitle = SortableTitle(title)
				} else {
					video.SortTitle = sortTitle
				}

				if g, err := videoFile.Genre(); err == nil && g != "" {
					video.Genre = &g
				}

				if d, err := videoFile.Description(); err == nil && d != "" {
					video.Description = &d
				}

				// XXX switch off mediakind?
				if seriesNameErr == nil && seriesName != "" {
					l.Mutex.Lock()
					series, ok := l.Series[seriesID]
					if !ok {
						series = &Series{
							Name: seriesName,
						}
						l.Series[seriesID] = series
					}

					season, ok := l.Seasons[seasonID]
					if !ok {
						season = &Season{
							Season: seasonNumber,
							Series: series,
						}
						l.Seasons[seasonID] = season
					}

					episode := video.Episode
					if episode == nil {
						episode = &Episode{
							Season:  season,
							Episode: episodeNumber,
							Video:   video, // XXX cyclic reference loop
						}
						video.Episode = episode
					}
					l.Mutex.Unlock()

					if sortSeriesName, err := videoFile.TVSortShowName(); err == nil && sortSeriesName != "" {
						series.SortName = sortSeriesName
					} else {
						series.SortName = SortableTitle(seriesName)
					}
				}

				renditionID := hashToStr(path)
				rendition := &Rendition{
					ID:      renditionID,
					Quality: qualityFromPath(path),
					Size:    int(info.Size()),
				}

				{
					l.Mutex.Lock()

					l.Metarenditions[renditionID] = &Metarendition{
						path,
					}

					if video.Renditions == nil {
						video.Renditions = &Renditions{}
					}
					video.Renditions.All = append(video.Renditions.All, rendition)
					l.Mutex.Unlock()
				}

				return nil
			}()
		})
	return err
}

func hashToStr(payload string) string {
	msg := sha256.Sum256([]byte(payload))
	return base64.StdEncoding.EncodeToString(msg[:])
}

func qualityFromPath(path string) *Quality {
	// HACK

	q := &Quality{
		VideoCodec: VideoCodecH264,
	}

	for {
		parent := filepath.Dir(path)
		if parent == "" || parent == "/" || parent == "." {
			break
		}
		pathElement := filepath.Base(parent)
		path = parent

		switch {
		case strings.Contains(pathElement, "--encoder=x265"):
			q.VideoCodec = VideoCodecH265
		}

		switch {
		case strings.Contains(pathElement, " 480p30"):
			q.Resolution = "480p"
		case strings.Contains(pathElement, " 720p30"):
			q.Resolution = "720p"
		case strings.Contains(pathElement, " 1080p30"):
			q.Resolution = "1080p"
		}

		switch {
		case strings.HasPrefix(pathElement, "Very Fast "):
			b := TranscodeBudgetVeryFast
			q.TranscodeBudget = &b
		case strings.HasPrefix(pathElement, "Fast "):
			b := TranscodeBudgetFast
			q.TranscodeBudget = &b
		case strings.HasPrefix(pathElement, "HQ "):
			b := TranscodeBudgetHq
			q.TranscodeBudget = &b
		case strings.HasPrefix(pathElement, "Super HQ "):
			b := TranscodeBudgetSuperHq
			q.TranscodeBudget = &b
		}
	}

	return q
}

func (l *Library) Episodes(series *SeriesFilter, season *SeasonFilter) ([]*Episode, error) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	var matches []*Episode
	for _, metavideo := range l.Metavideos {
		episode := metavideo.Video.Episode
		if episode == nil {
			continue
		}

		if season != nil && season.Season != nil {
			if episode.Season.Season != *season.Season {
				continue
			}
		}

		if series != nil && series.Name != nil {
			if episode.Season.Series.Name != *series.Name {
				continue
			}
		}

		matches = append(matches, episode)
	}

	return matches, nil
}
