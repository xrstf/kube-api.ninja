#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

rootDir="$(realpath .)"

get_version() {
  curl -sfL "https://dl.k8s.io/release/$1-$2.txt"
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
