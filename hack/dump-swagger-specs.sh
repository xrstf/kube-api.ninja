#!/usr/bin/env bash

set -euo pipefail

cd $(dirname $0)/..

make clean build

currentdev=1.29
INCLUDE_EOL=${INCLUDE_EOL:-false}

if ! $INCLUDE_EOL; then
  echo "Not including EOL releases, set INCLUDE_EOL=true to dump specs for all releases."
fi

today="$(date +'%Y-%m-%d')"

for releaseDir in data/releases/*; do
  release="$(basename "$releaseDir")"
  eolDate="$(cat "$releaseDir/eol.txt" 2>/dev/null || true)"

  if ! $INCLUDE_EOL && [ -n "$eolDate" ] && [[ "$eolDate" < "$today" ]]; then
    echo "Skipping release $release because it's end-of-life."
    continue
  fi

  echo "Dumping APIs for Kubernetes $release …"

  # allow to fetch the development branch before it was released
  branch="release-$release"
  if [[ "$currentdev" == "$release" ]]; then
    branch="master"
  fi

  wget --output-document swagger.json https://github.com/kubernetes/kubernetes/raw/$branch/api/openapi-spec/swagger.json

  _build/swaggerdumper \
    -swagger-file swagger.json \
    -kubernetes-version "$release.0" \
    > "data/releases/$release/api.json"
done

rm -f swagger.json
