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

type Library struct {
	Videos        map[string]*Video
	Series        map[string]*Series
	Seasons       map[string]*Season
	RenditionPath map[string]string
	Mutex         sync.Mutex
}

func NewLibrary() *Library {
	return &Library{
		Videos:        make(map[string]*Video),
		Series:        make(map[string]*Series),
		Seasons:       make(map[string]*Season),
		RenditionPath: make(map[string]string),
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

				id := idFromTitleAndYear(title, releaseYear)
				l.Mutex.Lock()
				video, ok := l.Videos[id]
				if !ok {
					video = &Video{
						ID:          id,
						Title:       title,
						ReleaseYear: releaseYear,
					}
					l.Videos[id] = video
				}
				l.Mutex.Unlock()

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
				if seriesName, err := videoFile.TVShowName(); err == nil && seriesName != "" {
					seasonNumber, err := videoFile.TVSeason()
					if err != nil {
						// HACK missing or zero season looks funny
						seasonNumber = 1
					}

					episodeNumber, _ := videoFile.TVEpisode()

					seasonID := idFromSeriesAndSeason(seriesName, seasonNumber)
					seriesID := hashToStr(seriesName)

					l.Mutex.Lock()
					series, ok := l.Series[seriesID]
					if !ok {
						series = &Series{
							ID:   seriesID,
							Name: seriesName,
						}
						l.Series[seriesID] = series
					}

					season, ok := l.Seasons[seasonID]
					if !ok {
						season = &Season{
							ID:     seasonID,
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
							Video:   video, // XXX circular reference loop
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

					l.RenditionPath[renditionID] = path
					// XXX HACK
					l.RenditionPath[id] = path

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

func idFromTitleAndYear(title string, releaseYear int) string {
	key := fmt.Sprintf("%s (%d)", title, releaseYear)
	return hashToStr(key)
}

func idFromSeriesAndSeason(series string, season int) string {
	key := fmt.Sprintf("%s (%d)", series, season)
	return hashToStr(key)
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
