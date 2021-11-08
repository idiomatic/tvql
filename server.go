package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/idiomatic/tvql/graph"
	"github.com/idiomatic/tvql/graph/generated"
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

	http.Handle("/video/",
		http.StripPrefix("/video",
			http.FileServer(http.Dir(root))))
	base, err := url.Parse(fmt.Sprintf("http://localhost:%s/video/", port))
	if err != nil {
		log.Panic(err)
	}

	resolver := graph.NewResolver()
	go func() {
		err := resolver.Survey(root, *base)
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
