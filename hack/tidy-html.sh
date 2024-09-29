#!/usr/bin/env bash

set -euo pipefail
cd $(dirname $0)/..

cd public
for file in *.html; do
  tidy -config ../tidy.conf -modify "$file"

  # trim trailing whitespace added by tidy
  # cf. https://github.com/htacg/tidy-html5/issues/523
  sed -i 's/[[:blank:]]\+$//g' "$file"
done
