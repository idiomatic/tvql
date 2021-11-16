package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/idiomatic/tvql/graph"
	"github.com/idiomatic/tvql/graph/generated"
	"github.com/idiomatic/tvql/graph/model"
)

const (
	defaultPort = "8080"
	defaultRoot = "/Users/brian/Review/Video"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	root := os.Getenv("ROOT")
	if root == "" {
		root = defaultRoot
	}

	base, err := url.Parse(fmt.Sprintf("http://localhost:%s/", port))
	if err != nil {
		log.Panic(err)
	}
	videoPath := &url.URL{Path: "/video/"}
	videoBase := base.ResolveReference(videoPath)
	artworkPath := &url.URL{Path: "/artwork/"}
	artworkBase := base.ResolveReference(artworkPath)

	library := model.NewLibrary()

	http.Handle(videoBase.Path,
		http.StripPrefix(videoBase.Path,
			InternalRedirect(func(u *url.URL) {
				library.Mutex.Lock()
				defer library.Mutex.Unlock()
				absPath := library.RenditionPath[u.Path]
				u.Path = strings.TrimPrefix(absPath, root)
			}, http.FileServer(http.Dir(root)))))

	resolver := graph.NewResolver(library, videoBase, artworkBase)

	http.Handle("/artwork/",
		http.StripPrefix("/artwork/",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id := r.URL.Path
				q := r.URL.Query()

				geometry := model.GeometryFilter{
					Width:  Atoiptr(q.Get("width")),
					Height: Atoiptr(q.Get("height")),
				}

				artwork, err := library.GetArtwork(id)
				if err != nil {
					http.Error(w, "artwork retrieval failed", http.StatusInternalServerError)
				}

				artwork, err = model.ResizeArtwork(artwork, &geometry)
				if err != nil {
					http.Error(w, "artwork resize failed", http.StatusInternalServerError)
				}

				w.Header().Add("Content-type", "image/jpeg")
				w.Write(artwork)
			})))

	go func() {
		err := library.Survey(root)
		if err != nil {
			log.Panic(err)
		}
	}()

	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func Atoiptr(s string) *int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &value
}

func InternalRedirect(fn func(u *url.URL), h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		fn(r2.URL)
		h.ServeHTTP(w, r2)
	})
}
