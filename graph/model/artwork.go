package model

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/disintegration/imaging"
	"github.com/idiomatic/tvql/metadata/mp4"
	"github.com/sunfish-shogi/bufseekio"
)

type Artwork struct {
	ID string
}

// XXX video id or rendition id?
func (l *Library) GetArtwork(id string) ([]byte, error) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	metavideo, ok := l.Metavideos[id]
	if !ok {
		return nil, fmt.Errorf("video not found")
	}

	path := metavideo.Path

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bufferedFile := bufseekio.NewReadSeeker(file, 1024, 4)

	videoFile := mp4.NewFile(bufferedFile)

	artwork, err := videoFile.CoverArt()
	if err != nil {
		return nil, err
	}

	return artwork, nil
}

func ResizeArtwork(artwork []byte, geometry *GeometryFilter) ([]byte, error) {
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
