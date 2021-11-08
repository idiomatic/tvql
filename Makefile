run:
	go run ./server.go

generate:
	go run github.com/99designs/gqlgen generate .

deps:
	go get github.com/99designs/gqlgen
	go run github.com/99designs/gqlgen init
