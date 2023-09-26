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

cd data/releases/
for release in *; do
  (
    cd "$release"

    # this can be an empty string if there is no stable version yet
    newVersion=$(get_version stable "$release" || true)

    # if we there is no new stable version, check unstable
    if [ -z "$newVersion" ]; then
      newVersion=$(get_version latest "$release")
    fi

    if [ -n "$newVersion" ]; then
      # trim leading v
      echo "${newVersion#v}" > latest.txt
    fi
  )
done
