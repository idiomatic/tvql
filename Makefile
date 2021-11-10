TARGET_ARCH=arm64
TARGET_HOST=marx
TARGET_OS=linux
TARGET_EXE=server-$(TARGET_OS)-$(TARGET_ARCH)

run:
	go run ./server.go

generate:
	go run github.com/99designs/gqlgen generate .

deps:
	go get github.com/99designs/gqlgen
	go run github.com/99designs/gqlgen init

deploy:
	env GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o $(TARGET_EXE) && scp $(TARGET_EXE) $(TARGET_HOST):
