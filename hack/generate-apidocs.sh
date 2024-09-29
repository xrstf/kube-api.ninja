#!/usr/bin/env bash

set -euo pipefail

cd $(dirname $0)/..

today="$(date +'%Y-%m-%d')"
INCLUDE_EOL=${INCLUDE_EOL:-false}

if ! $INCLUDE_EOL; then
  echo "Not including EOL releases, set INCLUDE_EOL=true to dump specs for all releases."
fi

if [ -z "${RELEASES:-}" ]; then
  RELEASES=""

  for release in $(ls data/releases | sort --version-sort); do
    releaseDir="data/releases/$release"
    eolDate="$(cat "$releaseDir/eol.txt" 2>/dev/null || true)"

    if ! $INCLUDE_EOL && [ -n "$eolDate" ] && [[ "$eolDate" < "$today" ]]; then
      echo "Skipping release $release because it's end-of-life."
      continue
    fi

    if [ ! -f "$releaseDir/apidocs.yaml" ]; then
      echo "Skipping release $release because it has no apidocs.yaml file."
      continue
    fi

    RELEASES="$RELEASES $release"
  done
fi

make build

for release in $RELEASES; do
  (
    releaseDir="public/apidocs/$release"

    set -x

    _build/apidocs \
      -kubernetes-release "$release" \
      -build-dir "$releaseDir"

    # warnings make tidy always exit with 1, there's no way to disable that
    # cf. https://github.com/htacg/tidy-html5/issues/1071
    tidy \
      -config public/apidocs/tidy.conf \
      -modify \
      -quiet \
      "$releaseDir/index.html" || true

    # trim trailing whitespace added by tidy
    # cf. https://github.com/htacg/tidy-html5/issues/523
    sed -i 's/[[:blank:]]\+$//g' "$releaseDir/index.html"
  )
done
