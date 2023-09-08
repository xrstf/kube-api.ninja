#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

docker run \
  --rm \
  -e "RELEASES=1.11 1.12 1.13 1.14 1.15 1.16 1.17 1.18 1.19 1.20 1.21 1.23 1.24 1.25 1.26 1.27 1.28 1.29" \
  -v "$(realpath public/apidocs/):/output" \
  kubernetes-apidocs:latest /make.sh
