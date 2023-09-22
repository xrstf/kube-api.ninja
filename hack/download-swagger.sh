#!/usr/bin/env bash

set -euo pipefail

cd $(dirname $0)/..

RELEASE="${1:-}"
if [ -z "$RELEASE" ]; then
  echo "Usage: hack/download-swagger.sh 1.42"
  exit 1
fi

releaseDir="data/releases/$RELEASE"
mkdir -p "$releaseDir"

# allow to fetch the development branch before it was released
currentdev="$(cat "data/kubernetes-master.txt")"
branch="release-$RELEASE"
if [[ "$currentdev" == "$RELEASE" ]]; then
  branch="master"
fi

wget --output-document "$releaseDir/swagger.json" https://github.com/kubernetes/kubernetes/raw/$branch/api/openapi-spec/swagger.json
