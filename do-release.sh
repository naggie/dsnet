#!/bin/sh
set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <tag>"
    echo "Release version required as argument"
    exit 1
fi

VERSION="$1"
GIT_COMMIT=$(git rev-list -1 HEAD)
BUILD_DATE=$(date)

RELEASE_FILE=RELEASE.md

export CGO_ENABLED=0
export GOOS=linux

LDFLAGS="-s -w \
    -X \"github.com/naggie/dsnet.GIT_COMMIT=$GIT_COMMIT\" \
    -X \"github.com/naggie/dsnet.VERSION=$VERSION\" \
    -X \"github.com/naggie/dsnet.BUILD_DATE=$BUILD_DATE\"\
"

# check tag starts with v
if [ "${VERSION:0:1}" != "v" ]; then
    echo "Tag must start with v"
    exit 1
fi

nvim "+ normal G $" $RELEASE_FILE

# build
mkdir -p dist

export GOOS=linux
export CGO_ENABLED=0

GOARCH=arm GOARM=5 go build -ldflags="$LDFLAGS" -o dist/dsnet cmd/root.go
# upx -q dsnet
mv dist/dsnet dist/dsnet-linux-arm5

GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/dsnet cmd/root.go
# upx -q dsnet
mv dist/dsnet dist/dsnet-linux-arm64

GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/dsnet cmd/root.go
# upx -q dsnet
mv dist/dsnet dist/dsnet-linux-amd64

# github.com/cli/cli
# https://github.com/cli/cli/releases/download/v2.15.0/gh_2.15.0_linux_amd64.deb
# do: gh auth login
gh release create \
    --title $VERSION \
    --notes-file $RELEASE_FILE \
    --draft \
    $VERSION \
    dist/dsnet-linux-arm5#"dsnet linux-arm5" \
    dist/dsnet-linux-arm64#"dsnet linux-arm64" \
    dist/dsnet-linux-amd64#"dsnet linux-amd64" \
