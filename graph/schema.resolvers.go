package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

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
	var matches []*model.Video

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if contributor != nil {
		panic(fmt.Errorf("not implemented"))
	}

	skipping := (paginate != nil && paginate.After != nil)
	for _, video := range r.videos {
		if title != nil && video.Title != *title {
			continue
		}

		if !skipping {
			matches = append(matches, video)
		}
		if paginate != nil {
			// skip until after client specified video found
			skipping = skipping && video.ID != *paginate.After

			// limit number of matches
			if paginate.First != nil && len(matches) >= *paginate.First {
				break
			}
		}
	}

	// XXX skip videos without suitable renditions

	return matches, nil
}

func (r *queryResolver) Series(ctx context.Context, paginate *model.Paginate) ([]*model.Series, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Episodes(ctx context.Context, series *model.SeriesFilter) ([]*model.Episode, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *videoResolver) Artwork(ctx context.Context, obj *model.Video, geometry *model.GeometryFilter) (*string, error) {
	return r.resizeArtwork(ctx, obj, geometry)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Video returns generated.VideoResolver implementation.
func (r *Resolver) Video() generated.VideoResolver { return &videoResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type videoResolver struct{ *Resolver }
