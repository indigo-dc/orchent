#!/bin/bash

GO=`which go`
PATH_TO_SCRIPT=`readlink -f ${0}`
PATH_TO_FOLDER=`dirname "$PATH_TO_SCRIPT"`

if [ "x$GO" == "x" ]; then
    echo "go missing, please install go 1.5 or newer"
    exit 1
fi

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
echo -n "building orchent ... "
go build -o orchent ${GOPATH}/orchent.go
echo "done"
