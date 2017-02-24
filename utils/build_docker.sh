#!/bin/bash

PATH_TO_SCRIPT=`readlink -f ${0}`
PATH_TO_FOLDER=`dirname "$PATH_TO_SCRIPT"`
PATH_TO_REPO=`cd "${PATH_TO_FOLDER}/.." && pwd -P`

DOCKERFILE="$PATH_TO_FOLDER/Dockerfile"
ORCHENT="$PATH_TO_REPO/orchent"

cd $REPO_PATH
pwd
echo " "
echo " building orchent ..."
./utils/compile.sh
echo " "
mkdir -p /tmp/orchent_docker/
cp $DOCKERFILE /tmp/orchent_docker/
cp $ORCHENT /tmp/orchent_docker/
cd /tmp/orchent_docker/
ORCHENT_VERSION=`./orchent --version 2>&1`
ORCHENT_TAG="orchent:$ORCHENT_VERSION"
docker image rm "$ORCHENT_TAG"
docker build -t "$ORCHENT_TAG" .
cd $PATH_TO_REPO
rm -rf /tmp/orchent_docker/
docker image save --output "orchent_container_${ORCHENT_VERSION}.tgz" "$ORCHENT_TAG"
echo "done"
