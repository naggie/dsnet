.PHONY: all build

all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -ldflags="-s -w" -o dist/dsnet ./cmd/dsnet.go
	upx dist/dsnet

update_deps:
	# `go mod vendor` initialises vendoring system
	go get
	go mod vendor
	git add -f vendor
