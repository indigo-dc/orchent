#!/bin/bash

realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}

GO=`which go`
REALPATH=`which realpath`
if [ "x$GO" == "x" ]; then
    echo "go missing, please install go 1.5 or newer"
    exit 1
fi

PATH_TO_SCRIPT=`realpath ${0}`
PATH_TO_FOLDER=`dirname "$PATH_TO_SCRIPT"`

VERSION=`go version`
GOPATH=`cd "${PATH_TO_FOLDER}/.." && pwd -P`
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
go get github.com/indigo-dc/liboidcagent-go/liboidcagent
echo "done"
echo -n "building orchent ... "
go build -o orchent ${GOPATH}/orchent.go
echo "done"
