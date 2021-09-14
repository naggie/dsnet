.PHONY: all build compile quick clean

all: build

clean:
	@rm -r dist

compile:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -ldflags="-s -w" -o dist/dsnet ./cmd/main.go

build: compile
	upx dist/dsnet

quick: compile

update_deps:
	# `go mod vendor` initialises vendoring system
	go get
	go mod vendor
	git add -f vendor

