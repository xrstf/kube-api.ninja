#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

mkdir -p data

make clean build

for release in 1.14 1.15 1.16 1.17 1.18 1.19 1.20 1.21 1.22 1.23 1.24 1.25 1.26 1.27 1.28; do
  echo "Dumping APIs for Kubernetes $release …"

  wget --quiet --output-document swagger.json https://github.com/kubernetes/kubernetes/raw/release-$release/api/openapi-spec/swagger.json

  _build/swaggerdumper \
    -swagger-file swagger.json \
    -kubernetes-version "$release.0" \
    > "data/release-$release-swagger.json"
done
