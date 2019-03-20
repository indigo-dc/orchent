#!/bin/bash

DOCKER_REPO=${1:-orchent}

GO=`which go`
REALPATH=`which realpath`
if [ "x$GO" == "x" ]; then
    echo "go missing, please install go 1.5 or newer"
    exit 1
fi

if [ "x$REALPATH" == "x" ]; then
    echo "realpath missing, please install it"
    exit 1
fi

PATH_TO_SCRIPT=`realpath ${0}`
PATH_TO_FOLDER=`dirname "$PATH_TO_SCRIPT"`
PATH_TO_REPO=`cd "${PATH_TO_FOLDER}/.." && pwd -P`

DOCKERFILE="$PATH_TO_FOLDER/Dockerfile"
ORCHENT="$PATH_TO_REPO/orchent"

cd $PATH_TO_REPO
echo " "
echo " building orchent ..."

VERSION=`go version`
GOPATH=`cd "${PATH_TO_FOLDER}/.." && pwd -P`

echo "    cleaning ..."
pwd
rm orchent
rm orchent_container_*.tgz
echo " "
echo "running the build with '$VERSION', please include in issue reports"
echo " "
export "GOPATH=${GOPATH}"
echo "fetiching:"
echo -n "  kingpin ... "
go get gopkg.in/alecthomas/kingpin.v2
echo "done"
echo -n "  sling ... "
go get github.com/dghubble/sling
echo "done"
echo -n "  go-config ... "
go get github.com/zpatrick/go-config
echo "done"
echo -n "  liboidcagent ... "
go get github.com/zachmann/liboidcagent-go/liboidcagent
echo "done"
echo -n "building orchent ... "
CGO_ENABLED=0 GOOS=linux go build -a -v -o $ORCHENT ${GOPATH}/orchent.go
echo "done"

echo "building docker ... "
mkdir -p /tmp/orchent_docker/
cp $DOCKERFILE /tmp/orchent_docker/
cp $ORCHENT /tmp/orchent_docker/
cp /etc/ssl/certs/ca-certificates.crt /tmp/orchent_docker/
cd /tmp/orchent_docker/
ORCHENT_VERSION=`./orchent --version 2>&1`
ORCHENT_TAG="$DOCKER_REPO:$ORCHENT_VERSION"
ORCHENT_DOCKER="$PATH_TO_REPO/orchent_container_${ORCHENT_VERSION}.tar"
docker image rm -f "$ORCHENT_TAG"
docker build --no-cache -t "$ORCHENT_TAG" .
cd $PATH_TO_REPO
rm -rf /tmp/orchent_docker/
docker save --output "$ORCHENT_DOCKER" "$ORCHENT_TAG"
echo "done"

echo " "
echo " "
echo " checking image "
docker image rm -f "$ORCHENT_TAG"
docker images -a
docker load --input "$ORCHENT_DOCKER"
docker run  --rm "$ORCHENT_TAG" --version
docker images -a
echo " done "
