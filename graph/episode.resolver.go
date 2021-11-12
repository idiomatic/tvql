package graph

import (
	"context"

	"github.com/idiomatic/tvql/graph/model"
)

func (r *seriesResolver) episodes(ctx context.Context, obj *model.Series) ([]*model.Episode, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.videos {
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.videos {
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.videos {
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
