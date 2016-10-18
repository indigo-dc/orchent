#!/bin/bash

my_dir="$(dirname "$0")"
source $my_dir/package_config.sh

mkdir -p packaging/debian/orchent/usr/bin
cp orchent packaging/debian/orchent/usr/bin

#  adjust the config files
mkdir -p packaging/debian/orchent/DEBIAN
cat packaging/debian/conf/control | ./scripts/mo > packaging/debian/orchent/DEBIAN/control
cat packaging/debian/conf/postinst | ./scripts/mo > packaging/debian/orchent/DEBIAN/postinst

dpkg --build packaging/debian/orchent/

mv packaging/debian/orchent.deb packaging/orchent-$VERSION-1_amd64.deb
