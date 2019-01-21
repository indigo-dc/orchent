#!/bin/bash


CURRENT_DIR=`pwd`
cd `dirname $0`
cd ..
SOURCE_DIR=`pwd`
VERSION=`git describe --abbrev=0 --tags`
source ./utils/prepare_system.sh

# expecting to be in $GOPATH/src/github.com/indigo-dc/orchent
mkdir -p $GOPATH/src/github.com/indigo-dc/orchent
cd $GOPATH/src/github.com/indigo-dc/orchent
pwd
cp -v $SOURCE_DIR/* .
glide install
cp -r vendor/* "$GOPATH/src"
rm -fr pkg-build/*
mkdir -p pkg-build/amd64
GOOS=linux GOARCH=amd64 go build -o build/amd64/orchent orchent.go
case $DISTRIBUTION in
    debian)
        GO_BIN="go-bin-deb"
        PKG="orchent_${VERSION}_amd64.deb"
        PFLAG="-w"
        ;;
    centos)
        GO_BIN="go-bin-rpm"
        PKG="orchent-$VERSION.el7.centos.x86_64.rpm"
        PFLAG="-b"
        ;;
esac
$GO_BIN generate -a amd64 --version ${VERSION} ${PFLAG} pkg-build/amd64 -o ${PKG}

echo " "
echo " "
echo " *** DONE ***"
