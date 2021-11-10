package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"sort"

	"github.com/idiomatic/tvql/graph/generated"
	"github.com/idiomatic/tvql/graph/model"
)

func (r *queryResolver) Video(ctx context.Context, id string) (*model.Video, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	video, ok := r.videos[id]
	if !ok {
		return nil, fmt.Errorf("video not found")
	}

	return video, nil
}

func (r *queryResolver) Videos(ctx context.Context, paginate *model.Paginate, title *string, contributor *model.ContributorFilter) ([]*model.Video, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if contributor != nil {
		panic(fmt.Errorf("not implemented"))
	}

	var matches []*model.Video
	for _, video := range r.videos {
		if title != nil && video.Title != *title {
			continue
		}

		matches = append(matches, video)
	}

	sort.Sort(model.ByVideoTitle(matches))

	if paginate != nil && paginate.After != nil {
		for i, video := range matches {
			if video.ID == *paginate.After {
				matches = matches[i+1:]
				break
			}
		}
	}

	if paginate != nil && paginate.First != nil && len(matches) > *paginate.First {
		matches = matches[:*paginate.First]
	}

	return matches, nil
}

func (r *queryResolver) Series(ctx context.Context, paginate *model.Paginate) ([]*model.Series, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Series
	for _, series := range r.series {
		matches = append(matches, series)
	}

	sort.Sort(model.BySeriesName(matches))

	if paginate != nil && paginate.After != nil {
		for i, series := range matches {
			if series.ID == *paginate.After {
				matches = matches[i+1:]
				break
			}
		}
	}

	if paginate != nil && paginate.First != nil && len(matches) > *paginate.First {
		matches = matches[:*paginate.First]
	}

	return matches, nil
}

func (r *queryResolver) Seasons(ctx context.Context, series *model.SeriesFilter) ([]*model.Season, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Season
	for _, season := range r.seasons {
		if series != nil {
			if season.Series.ID != *series.ID {
				continue
			}
		}

		matches = append(matches, season)
	}

	sort.Sort(model.BySeason(matches))

	return matches, nil
}

func (r *queryResolver) Episodes(ctx context.Context, season *model.SeasonFilter) ([]*model.Episode, error) {
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

		if season != nil && season.Series != nil {
			if season.Series.ID != nil && episode.Season.Series.ID != *season.Series.ID {
				continue
			}

			if season.Series.Name != nil && episode.Season.Series.Name != *season.Series.Name {
				continue
			}
		}

		matches = append(matches, episode)
	}

	sort.Sort(model.ByEpisode(matches))

	return matches, nil
}

func (r *seasonResolver) Episodes(ctx context.Context, obj *model.Season) ([]*model.Episode, error) {
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

	sort.Sort(model.ByEpisode(matches))

	return matches, nil
}

func (r *seriesResolver) Seasons(ctx context.Context, obj *model.Series) ([]*model.Season, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Season
	for _, season := range r.seasons {
		if obj != nil {
			if season.Series.ID != obj.ID {
				continue
			}
		}

		matches = append(matches, season)
	}

	sort.Sort(model.BySeason(matches))

	return matches, nil
}

func (r *seriesResolver) Episodes(ctx context.Context, obj *model.Series) ([]*model.Episode, error) {
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

	sort.Sort(model.ByEpisode(matches))

	return matches, nil
}

func (r *videoResolver) Renditions(ctx context.Context, obj *model.Video, quality *model.QualityFilter) ([]*model.Rendition, error) {
	// XXX filter on quality
	return obj.Renditions, nil
}

func (r *videoResolver) Artwork(ctx context.Context, obj *model.Video, geometry *model.GeometryFilter) (*string, error) {
	return r.resizeArtwork(ctx, obj, geometry)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Season returns generated.SeasonResolver implementation.
func (r *Resolver) Season() generated.SeasonResolver { return &seasonResolver{r} }

// Series returns generated.SeriesResolver implementation.
func (r *Resolver) Series() generated.SeriesResolver { return &seriesResolver{r} }

// Video returns generated.VideoResolver implementation.
func (r *Resolver) Video() generated.VideoResolver { return &videoResolver{r} }

type queryResolver struct{ *Resolver }
type seasonResolver struct{ *Resolver }
type seriesResolver struct{ *Resolver }
type videoResolver struct{ *Resolver }
