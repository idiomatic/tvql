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

func (r *mutationResolver) SetTomatometer(ctx context.Context, id string, tomatometer int) (*model.Video, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Video(ctx context.Context, id string) (*model.Video, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, video := range r.videos {
		if video.ID == id {
			return video, nil
		}
	}

	return nil, fmt.Errorf("video not found")
}

func (r *queryResolver) Videos(ctx context.Context, paginate *model.Paginate, title *string, contributor *model.ContributorFilter) ([]*model.Video, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if contributor != nil {
		panic(fmt.Errorf("not implemented"))
	}

	var matches []*model.Video
	for _, video := range r.videos {
		// filter by title
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

	// XXX skip videos without suitable renditions

	return matches, nil
}

func (r *queryResolver) Series(ctx context.Context, paginate *model.Paginate) ([]*model.Series, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// unique series
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

func (r *queryResolver) Episodes(ctx context.Context, series *model.SeriesFilter) ([]*model.Episode, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.videos {
		if video.Episode == nil {
			continue
		}
		if series != nil {
			if series.ID != nil && video.Episode.Series.ID != *series.ID {
				continue
			}
			if series.Name != nil && video.Episode.Series.Name != *series.Name {
				continue
			}
		}

		matches = append(matches, video.Episode)
	}

	return matches, nil
}

func (r *seriesResolver) Episodes(ctx context.Context, obj *model.Series) ([]*model.Episode, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var matches []*model.Episode
	for _, video := range r.videos {
		if video.Episode == nil {
			continue
		}
		if video.Episode.Series != obj {
			continue
		}

		matches = append(matches, video.Episode)
	}

	return matches, nil
}

func (r *videoResolver) Renditions(ctx context.Context, obj *model.Video, quality *model.QualityFilter) ([]*model.Rendition, error) {
	// XXX filter on quality
	return obj.Renditions, nil
}

func (r *videoResolver) Artwork(ctx context.Context, obj *model.Video, geometry *model.GeometryFilter) (*string, error) {
	return r.resizeArtwork(ctx, obj, geometry)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Series returns generated.SeriesResolver implementation.
func (r *Resolver) Series() generated.SeriesResolver { return &seriesResolver{r} }

// Video returns generated.VideoResolver implementation.
func (r *Resolver) Video() generated.VideoResolver { return &videoResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type seriesResolver struct{ *Resolver }
type videoResolver struct{ *Resolver }
