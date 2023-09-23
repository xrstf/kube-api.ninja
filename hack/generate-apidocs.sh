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

  for releaseDir in data/releases/*; do
    eolDate="$(cat "$releaseDir/eol.txt" 2>/dev/null || true)"
    release="$(basename $releaseDir)"

    if ! $INCLUDE_EOL && [ -n "$eolDate" ] && [[ "$eolDate" < "$today" ]]; then
      echo "Skipping release $release because it's end-of-life."
      continue
    fi

    if [ -f "$releaseDir/skipapidocs" ]; then
      echo "Skipping release $release because it has a skipapidocs file."
      continue
    fi

    RELEASES="$RELEASES $release"
  done
fi

docker run \
  --rm \
  -it \
  -e "RELEASES=$RELEASES" \
  -v "$(realpath public/apidocs/):/output" \
  kubernetes-apidocs:latest /make.sh

cd public/apidocs/
for release in $RELEASES; do
  cd "$release"

  sed -i 's#href="favicon.ico" type="image/vnd.microsoft.icon"#href="../static/images/favicon.png" type="image/png"#g' index.html
  sed -i 's#"/css/#"../static/css/#g' index.html
  sed -i 's#"/js/#"../static/js/#g' index.html

  cd ..
done
