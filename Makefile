.PHONY: all build compile quick clean

all: compile

clean:
	@rm -r dist

compile:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o dist/dsnet ./cmd/root.go

build: compile
	upx dist/dsnet

quick: compile

update_deps:
	go get

