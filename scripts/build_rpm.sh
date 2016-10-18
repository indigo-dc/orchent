#!/bin/bash

my_dir="$(dirname "$0")"
source $my_dir/package_config.sh

mkdir -p $RPM_DIR/SOURCES
cp orchent $RPM_DIR/SOURCES

#  adjust the config files
mkdir -p packaging/rpm/orchent/SPECS
cat packaging/rpm/conf/orchent.spec | ./scripts/mo > $RPM_DIR/SPECS/orchent.spec

rpmbuild --define "_topdir ${RPM_DIR}" -ba $RPM_DIR/SPECS/orchent.spec

mv $RPM_DIR/RPMS/x86_64/orchent-*.rpm packaging/
