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
	//videos  []*model.Video
	videos  map[string]*model.Video
	series  map[string]*model.Series
	artwork map[string][]byte
	mutex   sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
		videos:  make(map[string]*model.Video),
		series:  make(map[string]*model.Series),
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

			video, err := func() (*model.Video, error) {
				file, err := os.Open(path)
				if err != nil {
					return nil, err
				}
				defer file.Close()

				bufferedFile := bufseekio.NewReadSeeker(file, 1024, 4)

				videoFile := mp4.NewFile(bufferedFile)

				title, err := videoFile.Title()
				if err != nil {
					title = strings.TrimSuffix(filepath.Base(path), ext)
				}

				releaseDate, err := videoFile.ReleaseDate()
				if err != nil {
					releaseDate = "1900"
				}
				releaseYear, _ := strconv.Atoi(releaseDate[:4])

				id := idFromTitleAndYear(title, releaseYear)
				r.mutex.Lock()
				video, ok := r.videos[id]
				if !ok {
					video = &model.Video{
						ID: id, Title: title, ReleaseYear: releaseYear,
					}
					r.videos[id] = video
				}
				r.mutex.Unlock()

				sortTitle, _ := videoFile.SortTitle()
				if sortTitle == "" {
					sortTitle = model.SortableTitle(title)
				}
				video.SortTitle = sortTitle

				if g, err := videoFile.Genre(); err == nil {
					video.Genre = &g
				}

				if d, err := videoFile.Description(); err == nil {
					video.Description = &d
				}

				// XXX memory intensive
				// XXX should out-of-process cache original and downsamples
				if a, err := videoFile.CoverArt(); err == nil {
					r.artwork[video.ID] = a
				}

				// XXX switch off mediakind?
				if seriesName, err := videoFile.TVShowName(); err == nil && seriesName != "" {
					seriesID := hashToStr(seriesName)
					sortSeriesName, _ := videoFile.TVSortShowName()
					if sortSeriesName != "" {
						sortSeriesName = model.SortableTitle(seriesName)
					}

					r.mutex.Lock()
					series, ok := r.series[seriesID]
					if !ok {
						series = &model.Series{
							ID:       seriesID,
							Name:     seriesName,
							SortName: sortSeriesName,
						}

						r.series[seriesID] = series
					}
					r.mutex.Unlock()

					season, _ := videoFile.TVSeason()
					episode, _ := videoFile.TVEpisode()

					video.Episode = &model.Episode{
						Series:  series,
						Season:  season,
						Episode: episode,
						Video:   video, // XXX circular reference loop
					}
				}

				return video, nil
			}()

			relativeURL, err := url.Parse(
				strings.TrimPrefix(
					strings.TrimPrefix(path, root),
					"/",
				),
			)
			if err != nil {
				return err
			}
			resolvedURL := base.ResolveReference(relativeURL)

			rendition := &model.Rendition{
				ID:      hashToStr(path),
				URL:     resolvedURL.String(),
				Size:    int(info.Size()),
				Quality: qualityFromPath(path),
			}

			r.mutex.Lock()
			video.Renditions = append(video.Renditions, rendition)
			r.mutex.Unlock()

			return nil
		})
	return err
}
