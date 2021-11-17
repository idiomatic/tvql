package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/idiomatic/tvql/graph/generated"
	"github.com/idiomatic/tvql/graph/model"
)

func (r *artworkResolver) URL(ctx context.Context, obj *model.Artwork, geometry *model.GeometryFilter) (string, error) {
	values := make(url.Values)
	if geometry != nil {
		if geometry.Width != nil {
			values.Set("width", strconv.Itoa(*geometry.Width))
		}
		if geometry.Height != nil {
			values.Set("height", strconv.Itoa(*geometry.Height))
		}
	}

	relativeURL := &url.URL{
		Path:     obj.ID,
		RawQuery: values.Encode(),
	}
	resolvedURL := r.artworkBase.ResolveReference(relativeURL)

	return resolvedURL.String(), nil
}

func (r *artworkResolver) Base64(ctx context.Context, obj *model.Artwork, geometry *model.GeometryFilter) (string, error) {
	artwork, err := r.library.GetArtwork(obj.ID)
	if err != nil {
		return "", err
	}

	artwork, err = model.ResizeArtwork(artwork, geometry)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(artwork), nil
}

func (r *queryResolver) Video(ctx context.Context, id string) (*model.Video, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	metavideo, ok := r.library.Metavideos[id]
	if !ok {
		return nil, fmt.Errorf("video not found")
	}

	return &metavideo.Video, nil
}

func (r *queryResolver) Videos(ctx context.Context, paginate *model.Paginate, title *string, contributor *model.ContributorFilter) ([]*model.Video, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	if contributor != nil {
		panic(fmt.Errorf("not implemented"))
	}

	var matches []*model.Video
	for _, metavideo := range r.library.Metavideos {
		if title != nil && metavideo.Video.Title != *title {
			continue
		}

		matches = append(matches, &metavideo.Video)
	}

	sort.Sort(model.ByVideoTitle(matches))

	if paginate != nil && paginate.After != nil {
		for i, video := range matches {
			if video.ID == *paginate.After {
				matches = matches[i+1:]
				break
			}
		}
		// XXX should a not-found After result in []?
	}

	if paginate != nil && paginate.First != nil && len(matches) > *paginate.First {
		matches = matches[:*paginate.First]
	}

	return matches, nil
}

func (r *queryResolver) Series(ctx context.Context, paginate *model.Paginate) ([]*model.Series, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	var matches []*model.Series
	for _, series := range r.library.Series {
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
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	var matches []*model.Season
	for _, season := range r.library.Seasons {
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

func (r *queryResolver) Episodes(ctx context.Context, series *model.SeriesFilter, season *model.SeasonFilter) ([]*model.Episode, error) {
	matches, err := r.episodes(ctx, series, season)
	if err != nil {
		return nil, err
	}

	sort.Sort(model.ByEpisode(matches))

	return matches, nil
}

func (r *queryResolver) EpisodeCount(ctx context.Context, series *model.SeriesFilter, season *model.SeasonFilter) (int, error) {
	matches, err := r.episodes(ctx, series, season)
	if err != nil {
		return 0, err
	}

	return len(matches), nil
}

func (r *renditionResolver) URL(ctx context.Context, obj *model.Rendition) (string, error) {
	_, ok := r.library.Metarenditions[obj.ID]
	if !ok {
		return "", fmt.Errorf("rendition not found")
	}

	relativeURL := &url.URL{Path: obj.ID}
	resolvedURL := r.videoBase.ResolveReference(relativeURL)

	return resolvedURL.String(), nil
}

func (r *renditionsResolver) Rendition(ctx context.Context, obj *model.Renditions, quality *model.QualityFilter) (*model.Rendition, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	for _, rendition := range obj.All {
		if quality != nil && rendition.Quality != nil {
			// XXX func (vc *VideoCodec) Matches(quality *QualityFilter) bool
			if quality.VideoCodec != nil && *quality.VideoCodec != rendition.Quality.VideoCodec {
				continue
			}
			if quality.Resolution != nil && *quality.Resolution != rendition.Quality.Resolution {
				continue
			}
		}
		return rendition, nil
	}
	return nil, nil
}

func (r *seasonResolver) Episodes(ctx context.Context, obj *model.Season) ([]*model.Episode, error) {
	matches, err := r.episodes(ctx, obj)
	if err != nil {
		return nil, err
	}

	sort.Sort(model.ByEpisode(matches))

	return matches, nil
}

func (r *seasonResolver) EpisodeCount(ctx context.Context, obj *model.Season) (int, error) {
	matches, err := r.episodes(ctx, obj)
	if err != nil {
		return 0, err
	}

	return len(matches), nil
}

func (r *seriesResolver) Seasons(ctx context.Context, obj *model.Series) ([]*model.Season, error) {
	r.library.Mutex.Lock()
	defer r.library.Mutex.Unlock()

	var matches []*model.Season
	for _, season := range r.library.Seasons {
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
	matches, err := r.episodes(ctx, obj)
	if err != nil {
		return nil, err
	}

	sort.Sort(model.ByEpisode(matches))

	return matches, nil
}

func (r *seriesResolver) EpisodeCount(ctx context.Context, obj *model.Series) (int, error) {
	matches, err := r.episodes(ctx, obj)
	if err != nil {
		return 0, err
	}

	return len(matches), nil
}

func (r *videoResolver) Artwork(ctx context.Context, obj *model.Video) (*model.Artwork, error) {
	if _, err := r.library.GetArtwork(obj.ID); err != nil {
		return nil, nil
	}
	return &model.Artwork{ID: obj.ID}, nil
}

// Artwork returns generated.ArtworkResolver implementation.
func (r *Resolver) Artwork() generated.ArtworkResolver { return &artworkResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Rendition returns generated.RenditionResolver implementation.
func (r *Resolver) Rendition() generated.RenditionResolver { return &renditionResolver{r} }

// Renditions returns generated.RenditionsResolver implementation.
func (r *Resolver) Renditions() generated.RenditionsResolver { return &renditionsResolver{r} }

// Season returns generated.SeasonResolver implementation.
func (r *Resolver) Season() generated.SeasonResolver { return &seasonResolver{r} }

// Series returns generated.SeriesResolver implementation.
func (r *Resolver) Series() generated.SeriesResolver { return &seriesResolver{r} }

// Video returns generated.VideoResolver implementation.
func (r *Resolver) Video() generated.VideoResolver { return &videoResolver{r} }

type artworkResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type renditionResolver struct{ *Resolver }
type renditionsResolver struct{ *Resolver }
type seasonResolver struct{ *Resolver }
type seriesResolver struct{ *Resolver }
type videoResolver struct{ *Resolver }
