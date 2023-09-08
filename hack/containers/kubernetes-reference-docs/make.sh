#!/usr/bin/env bash

set -e

cd /go/src/github.com/kubernetes/reference-docs

for release in $RELEASES; do
  echo "Generating API documentation for Kubernetes $release …"

  mkdir -p "/output/$release"

  K8S_RELEASE=$release make api
  cp gen-apidocs/build/index.html gen-apidocs/build/navData.js "/output/$release/"
done
