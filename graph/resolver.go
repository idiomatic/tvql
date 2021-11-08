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
	videos  []*model.Video
	artwork map[string][]byte
	mutex   sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
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
				var video *model.Video
				for _, v := range r.videos {
					if v.ID == id {
						video = v
						break
					}
				}
				if video == nil {
					video = &model.Video{
						ID: id, Title: title, ReleaseYear: releaseYear,
					}
					r.videos = append(r.videos, video)
				}
				r.mutex.Unlock()

				if st, err := videoFile.SortTitle(); err == nil {
					video.SortTitle = &st
				}

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
				if tvshow, err := videoFile.TVShow(); err == nil && tvshow != "" {
					video.Episode = &model.Episode{}
					video.Episode.Series = &model.Series{
						ID:   hashToStr(tvshow),
						Name: tvshow,
					}
					video.Episode.Season, _ = videoFile.TVSeason()
					video.Episode.Episode, _ = videoFile.TVEpisode()
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
				ID:   hashToStr(path),
				URL:  resolvedURL.String(),
				Size: int(info.Size()),
			}

			r.mutex.Lock()
			video.Renditions = append(video.Renditions, rendition)
			r.mutex.Unlock()

			return nil
		})
	return err
}
