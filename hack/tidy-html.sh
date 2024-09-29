#!/usr/bin/env bash

set -euo pipefail
cd $(dirname $0)/..

cd public
for file in *.html; do
  tidy -config ../tidy.conf -modify "$file"
done
