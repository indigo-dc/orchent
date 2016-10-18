#!/bin/bash

my_dir="$(dirname "$0")"
source $my_dir/package_config.sh

mkdir -p $DEB_DIR/usr/bin
cp orchent $DEB_DIR/usr/bin

#  adjust the config files
mkdir -p $DEB_DIR/DEBIAN
cat $DEB_DIR/../conf/control | ./scripts/mo > $DEB_DIR/DEBIAN/control
cat $DEB_DIR/../conf/postinst | ./scripts/mo > $DEB_DIR/DEBIAN/postinst

dpkg --build $DEB_DIR

mv $DEB_DIR/../orchent.deb packaging/orchent-$VERSION-1_amd64.deb
