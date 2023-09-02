#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

mkdir -p apidocs-tmp

# 1.22 and 1.24 are missing in upstream, but were available via archive.org
for release in 1.17 1.18 1.19 1.20 1.21 1.23 1.25 1.26 1.27 1.28; do
  echo "Generating API documentation for Kubernetes $release …"

  docker run \
    --rm \
    -e "K8S_RELEASE=$release" \
    -v "$(realpath apidocs-tmp):/output" \
    kubernetes-apidocs:latest \
    sh -c 'make api && cp gen-apidocs/build/index.html gen-apidocs/build/navData.js /output/'

  mkdir -p "public/apidocs/$release"
  mv -f apidocs-tmp/* "public/apidocs/$release"
done

rmdir apidocs-tmp
