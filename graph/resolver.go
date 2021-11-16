package graph

import (
	"context"
	"net/url"

	"github.com/idiomatic/tvql/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	library     *model.Library
	videoBase   *url.URL
	artworkBase *url.URL
}

func NewResolver(library *model.Library, videoBase *url.URL, artworkBase *url.URL) *Resolver {
	return &Resolver{
		library:     library,
		videoBase:   videoBase,
		artworkBase: artworkBase,
	}
}

// XXX move to model/library.go
func (r *seriesResolver) episodes(ctx context.Context, obj *model.Series) ([]*model.Episode, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.library.Videos {
		episode := video.Episode
		if episode == nil {
			continue
		}

		if obj != nil {
			if episode.Season.Series.ID != obj.ID {
				continue
			}
		}

		matches = append(matches, episode)
	}

	return matches, nil
}

func (r *queryResolver) episodes(ctx context.Context, series *model.SeriesFilter, season *model.SeasonFilter) ([]*model.Episode, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.library.Videos {
		episode := video.Episode
		if episode == nil {
			continue
		}

		if season != nil {
			if season.ID != nil && episode.Season.ID != *season.ID {
				continue
			}
		}

		if series != nil {
			if series.ID != nil && episode.Season.Series.ID != *series.ID {
				continue
			}

			if series.Name != nil && episode.Season.Series.Name != *series.Name {
				continue
			}
		}

		matches = append(matches, episode)
	}

	return matches, nil
}

func (r *seasonResolver) episodes(ctx context.Context, obj *model.Season) ([]*model.Episode, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.library.Videos {
		episode := video.Episode
		if episode == nil {
			continue
		}

		if obj != nil {
			if episode.Season.ID != obj.ID {
				continue
			}
		}

		matches = append(matches, episode)
	}

	return matches, nil
}
