.PHONY: all build compile quick

all: build

compile:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -ldflags="-s -w" -o dist/dsnet ./cmd/dsnet.go

build: compile
	upx dist/dsnet

quick: compile

update_deps:
	# `go mod vendor` initialises vendoring system
	go get
	go mod vendor
	git add -f vendor
