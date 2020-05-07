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

LDFLAGS="-s -w \
    -X \"github.com/naggie/dsnet.GIT_COMMIT=$GIT_COMMIT\" \
    -X \"github.com/naggie/dsnet.VERSION=$VERSION\" \
    -X \"github.com/naggie/dsnet.BUILD_DATE=$BUILD_DATE\"\
"

# get release information
if ! test -f $RELEASE_FILE || head -n 1 $RELEASE_FILE | grep -vq $VERSION; then
    # file doesn't exist or is for old version, replace
    printf "$VERSION\n\n\n" > $RELEASE_FILE
fi

vim "+ normal G $" $RELEASE_FILE


# build
mkdir -p dist

export GOOS=linux
export CGO_ENABLED=0

GOARCH=arm GOARM=5 go build -ldflags="$LDFLAGS" cmd/dsnet.go
# upx -q dsnet
mv dsnet dist/dsnet-linux-arm5

GOARCH=amd64 go build -ldflags="$LDFLAGS" cmd/dsnet.go
# upx -q dsnet
mv dsnet dist/dsnet-linux-amd64

hub release create \
    --draft \
    -a dist/dsnet-linux-arm5#"dsnet linux-arm5" \
    -a dist/dsnet-linux-amd64#"dsnet linux-amd64" \
    -F $RELEASE_FILE \
    $1
