#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

# favicon does not live in static/ because some ancient
# user agents still default to fetching /favicon.ico regardless
# what the meta tags say

# thank you, https://www.jvt.me/posts/2022/02/07/favicon-cli/
convert data/logo.png \
  \( -clone 0 -resize 16x16 \) \
  \( -clone 0 -resize 32x32 \) \
  \( -clone 0 -resize 48x48 \) \
  \( -clone 0 -resize 64x64 \) \
  -delete 0 \
  -colors 256 \
  public/favicon.ico
