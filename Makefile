.PHONY: all build

all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o dist/dsnet ./cmd/dsnet.go
	upx dist/dsnet
