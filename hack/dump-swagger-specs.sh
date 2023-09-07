#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

make clean build

currentdev=1.29

for release in 1.11 1.12 1.13 1.14 1.15 1.16 1.17 1.18 1.19 1.20 1.21 1.22 1.23 1.24 1.25 1.26 1.27 1.28 1.29; do
  echo "Dumping APIs for Kubernetes $release …"

  # allow to fetch the development branch before it was released
  branch="release-$release"
  if [[ "$currentdev" == "$release" ]]; then
    branch="master"
  fi

  wget --output-document swagger.json https://github.com/kubernetes/kubernetes/raw/$branch/api/openapi-spec/swagger.json

  mkdir -p "data/releases/$release"
  _build/swaggerdumper \
    -swagger-file swagger.json \
    -kubernetes-version "$release.0" \
    > "data/releases/$release/api.json"
done

rm swagger.json
