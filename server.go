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
	defaultBase = "http://localhost/"
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

	base, err := url.Parse(defaultBase)
	if err != nil {
		log.Panic(err)
	}
	if idx := strings.Index(base.Host, ":"); idx != -1 {
		base.Host = base.Host[:idx]
	}
	base.Host = fmt.Sprintf("%s:%s", base.Host, port)
	videoBase := base.ResolveReference(&url.URL{Path: "/video/"})
	artworkBase := base.ResolveReference(&url.URL{Path: "/artwork/"})

	library := model.NewLibrary()
	go func() {
		err := library.Survey(root)
		if err != nil {
			log.Panic(err)
		}
	}()

	http.Handle(videoBase.Path,
		http.StripPrefix(videoBase.Path,
			InternalRedirect(func(u *url.URL) {
				// map base64(video.id) to relative path
				library.Mutex.Lock()
				defer library.Mutex.Unlock()
				metarendition, ok := library.Metarenditions[u.Path]
				if ok {
					u.Path = strings.TrimPrefix(metarendition.Path, root)
				}
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

	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to %s for GraphQL playground", base.String())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Atoiptr returns a pointer to an int parsed from a string, else nil.
func Atoiptr(s string) *int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &value
}

// InternalRedirect is like StripPrefix except it allows full
// manipulation of the request URL.
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
