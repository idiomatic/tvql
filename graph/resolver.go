package graph

import (
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
