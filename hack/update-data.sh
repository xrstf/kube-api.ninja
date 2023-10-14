#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

rootDir="$(realpath .)"

get_version() {
  # Do not use dl.k8s.io, as it gives completely random results after new releases
  # come out, as the CDN is not synced globally and doing the exact same request
  # twice in a row can very easily return the wrong results.
  # See also https://github.com/kubernetes/k8s.io/issues/5755
  curl -sfL "https://storage.googleapis.com/kubernetes-release/release/$1-$2.txt"
}

rootDir="$(realpath .)"
hasBuiltTools=false

cd data/releases/
for release in $(ls | sort --version-sort); do
  echo "Checking $release…"

  lastVersion="$(cat "$release/latest.txt")"

  # this can be an empty string if there is no stable version yet
  newVersion=$(get_version stable "$release" || true)

  # if we there is no new stable version, check unstable
  if [ -z "$newVersion" ]; then
    newVersion=$(get_version latest "$release")
  fi

  if [ -n "$newVersion" ]; then
    # trim leading v
    newVersion="${newVersion#v}"

    echo "$newVersion" > "$release/latest.txt"

    if [ "$lastVersion" != "$newVersion" ]; then
      echo "  $lastVersion => $newVersion"

      if ! $hasBuiltTools; then
        make -C "$rootDir" clean build
        hasBuiltTools=true
      fi

      $rootDir/hack/download-swagger.sh "$release"
      $rootDir/_build/swaggerdumper -swagger-file "$release/swagger.json" -kubernetes-version "$release.0" > "$release/api.json"
    fi
  fi
done
