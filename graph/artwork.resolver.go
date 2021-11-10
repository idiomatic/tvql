package graph

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"

	"github.com/disintegration/imaging"
	"github.com/idiomatic/tvql/graph/model"
)

func (r *videoResolver) resizeArtwork(ctx context.Context, video *model.Video, geometry *model.GeometryFilter) (*string, error) {
	id := video.ID
	artwork, ok := r.artwork[id]
	if !ok {
		return nil, nil
	}

	artwork, err := resizeArtwork(artwork, geometry)
	artworkString := base64.StdEncoding.EncodeToString(artwork)
	return &artworkString, err
}

func resizeArtwork(artwork []byte, geometry *model.GeometryFilter) ([]byte, error) {
	if geometry == nil {
		return artwork, nil
	}
	if geometry.Width == nil && geometry.Height == nil {
		return artwork, nil
	}

	rawImage, _, err := image.Decode(bytes.NewBuffer(artwork))
	if err != nil {
		return nil, err
	}

	var height, width int
	size := rawImage.Bounds().Size()

	if geometry.Width == nil && geometry.Height != nil {
		height = *geometry.Height
		if height > size.Y {
			height = size.Y
		}
		width = height * size.X / size.Y
	} else if geometry.Height == nil && geometry.Width != nil {
		width = *geometry.Width
		if width > size.X {
			width = size.X
		}
		height = width * size.Y / size.X
	} else {
		width = *geometry.Width
		height = *geometry.Height
		if width > size.X {
			width = size.X
		}
		if height > size.Y {
			height = size.Y
		}
	}

	resizedImage := imaging.Resize(rawImage, width, height, imaging.Lanczos)

	resizedArtwork := bytes.NewBuffer(nil)
	if err := jpeg.Encode(resizedArtwork, resizedImage, nil); err != nil {
		return nil, err
	}

	return resizedArtwork.Bytes(), nil
}
