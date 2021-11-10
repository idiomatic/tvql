package graph

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/idiomatic/tvql/graph/model"
	"github.com/idiomatic/tvql/metadata/mp4"
	"github.com/sunfish-shogi/bufseekio"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	videos              map[string]*model.Video
	series              map[string]*model.Series
	seasons             map[string]*model.Season
	artwork             map[string][]byte
	ArtworkResizeHeight *int
	mutex               sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
		videos:  make(map[string]*model.Video),
		series:  make(map[string]*model.Series),
		seasons: make(map[string]*model.Season),
		artwork: make(map[string][]byte),
	}
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

func qualityFromPath(path string) *model.Quality {
	// HACK

	q := &model.Quality{
		VideoCodec: model.VideoCodecH264,
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
			q.VideoCodec = model.VideoCodecH265
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
			b := model.TranscodeBudgetVeryFast
			q.TranscodeBudget = &b
		case strings.HasPrefix(pathElement, "Fast "):
			b := model.TranscodeBudgetFast
			q.TranscodeBudget = &b
		case strings.HasPrefix(pathElement, "HQ "):
			b := model.TranscodeBudgetHq
			q.TranscodeBudget = &b
		case strings.HasPrefix(pathElement, "Super HQ "):
			b := model.TranscodeBudgetSuperHq
			q.TranscodeBudget = &b
		}
	}

	return q
}

func (r *Resolver) Survey(root string, base url.URL) error {
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
					title = strings.TrimSuffix(filepath.Base(path), ext)
				}

				releaseDate, err := videoFile.ReleaseDate()
				if err != nil || releaseDate == "" {
					releaseDate = "1900"
				}
				releaseYear, _ := strconv.Atoi(releaseDate[:4])

				id := idFromTitleAndYear(title, releaseYear)
				r.mutex.Lock()
				video, ok := r.videos[id]
				if !ok {
					video = &model.Video{
						ID:          id,
						Title:       title,
						ReleaseYear: releaseYear,
					}
					r.videos[id] = video
				}
				r.mutex.Unlock()

				if sortTitle, err := videoFile.SortTitle(); err != nil || sortTitle == "" {
					video.SortTitle = model.SortableTitle(title)
				} else {
					video.SortTitle = sortTitle
				}

				if g, err := videoFile.Genre(); err == nil && g != "" {
					video.Genre = &g
				}

				if d, err := videoFile.Description(); err == nil && d != "" {
					video.Description = &d
				}

				// XXX memory intensive
				// XXX should out-of-process cache original and downsamples
				if a, err := videoFile.CoverArt(); err == nil && len(a) > 0 {
					if r.ArtworkResizeHeight != nil {
						// HACK downsample now
						a, err = resizeArtwork(a, &model.GeometryFilter{Height: r.ArtworkResizeHeight})
						if err != nil {
							return err
						}
					}

					r.artwork[video.ID] = a
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

					r.mutex.Lock()
					series, ok := r.series[seriesID]
					if !ok {
						series = &model.Series{
							ID:   seriesID,
							Name: seriesName,
						}
						r.series[seriesID] = series
					}

					season, ok := r.seasons[seasonID]
					if !ok {
						season = &model.Season{
							ID:     seasonID,
							Season: seasonNumber,
							Series: series,
						}
						r.seasons[seasonID] = season
					}

					episode := video.Episode
					if episode == nil {
						episode = &model.Episode{
							Season:  season,
							Episode: episodeNumber,
							Video:   video, // XXX circular reference loop
						}
						video.Episode = episode
					}
					r.mutex.Unlock()

					if sortSeriesName, err := videoFile.TVSortShowName(); err == nil && sortSeriesName != "" {
						series.SortName = sortSeriesName
					} else {
						series.SortName = model.SortableTitle(seriesName)
					}
				}

				relativePath := strings.TrimPrefix(strings.TrimPrefix(path, root), "/")
				relativeURL := &url.URL{Path: relativePath}
				resolvedURL := base.ResolveReference(relativeURL)

				rendition := &model.Rendition{
					ID:      hashToStr(path),
					URL:     resolvedURL.String(),
					Size:    int(info.Size()),
					Quality: qualityFromPath(path),
				}

				{
					r.mutex.Lock()
					video.Renditions = append(video.Renditions, rendition)
					r.mutex.Unlock()
				}

				return nil
			}()
		})
	return err
}
