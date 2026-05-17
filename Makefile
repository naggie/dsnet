.PHONY: all build compile quick clean test cover cover-html lint

all: compile

clean:
	@rm -rf dist coverage.out coverage.html

compile:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o dist/dsnet ./cmd/root.go

build: compile
	upx dist/dsnet

quick: compile

test:
	CGO_ENABLED=0 go test ./...

lint:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-clean:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	CGO_ENABLED=0 go vet ./...

cover:
	CGO_ENABLED=0 go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "open coverage.html in a browser"

update_deps:
	go get

