#!/usr/bin/env bash

set -euo pipefail

cd $(dirname $0)/..

make clean build

currentdev="$(cat "data/kubernetes-master.txt")"
INCLUDE_EOL=${INCLUDE_EOL:-false}

if ! $INCLUDE_EOL; then
  echo "Not including EOL releases, set INCLUDE_EOL=true to dump specs for all releases."
fi

today="$(date +'%Y-%m-%d')"

for release in $(ls data/releases | sort --version-sort); do
  releaseDir="data/releases/$release"
  eolDate="$(cat "$releaseDir/eol.txt" 2>/dev/null || true)"

  if ! $INCLUDE_EOL && [ -n "$eolDate" ] && [[ "$eolDate" < "$today" ]]; then
    echo "Skipping release $release because it's end-of-life."
    continue
  fi

  if [ ! -f "$releaseDir/swagger.json" ]; then
    echo "Skipping release $release because it does not have a Swagger spec."
    continue
  fi

  echo "Dumping APIs for Kubernetes $release …"

  _build/swaggerdumper \
    -swagger-file "$releaseDir/swagger.json" \
    -kubernetes-version "$release.0" \
    > "$releaseDir/api.json"
done
