ARCH ?= darwin
REPO_VERSION := $$(git describe --abbrev=0 --tags)
BUILD_DATE := $$(date +%Y-%m-%d-%H:%M)
GIT_HASH := $$(git rev-parse --short HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-s -X main.Version=$(REPO_VERSION) -X main.GitCommit=$(GIT_HASH) -X main.BuildDate=$(BUILD_DATE)"
BINARY := ino
MAIN_PKG := github.com/ralreegorganon/ino/cmd/ino

DB_USER := ino
DB_PASSWORD := ino
DB_PORT_MIGRATION := 9432

INO_CONNECTION_STRING_LOCAL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_USER)?sslmode=disable
INO_CONNECTION_STRING_DOCKER := postgres://$(DB_USER):$(DB_PASSWORD)@db:5432/$(DB_USER)?sslmode=disable
INO_CONNECTION_STRING_MIGRATION_DOCKER := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT_MIGRATION)/$(DB_USER)?sslmode=disable

dep:
	dep ensure

build:
	go build -i -v -o build/bin/$(ARCH)/$(BINARY) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

run: build
	INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_LOCAL)" ./build/bin/$(ARCH)/$(BINARY)

install:
	go install $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

migrate:
	cd migrations/ && INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_LOCAL)" ./run-migrations

docker:
	GOOS=linux GOARCH=amd64 go build -o build/bin/linux/$(BINARY) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

run-docker: docker
	cd build/ && DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_DOCKER)" docker-compose -p ino rm -f ino
	DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_DOCKER)" docker-compose -f build/docker-compose.yml -p ino build
	DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_DOCKER)" docker-compose -f build/docker-compose.yml -p ino up -d

stop-docker:
	cd build/ && DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_DOCKER)" docker-compose -p ino stop

migrate-docker:
	cd migrations/ && INO_CONNECTION_STRING="$(INO_CONNECTION_STRING_MIGRATION_DOCKER)" ./run-migrations

clean:
	rm -rf build/bin/*

.PHONY: build install
