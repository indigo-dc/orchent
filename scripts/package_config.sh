#!/bin/bash

export VERSION="$(./orchent --version 2>&1 )"
export RPM_DIR=`pwd`/packaging/rpm/orchent
export DEB_DIR=`pwd`/packaging/deb/orchent


echo "exporting variables"
echo "  VERSION=$VERSION"
echo "  RPM_DIR=$RPM_DIR"
echo "  DEB_DIR=$DEB_DIR"
