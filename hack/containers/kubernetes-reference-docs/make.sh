#!/usr/bin/env bash

set -e

cd /go/src/github.com/kubernetes/reference-docs

for release in $RELEASES; do
  echo "Generating API documentation for Kubernetes $release …"

  mkdir -p "/output/$release"

  K8S_RELEASE=$release make api

  # beautify HTML / make validations easier by preventing
  # loads of info messages on https://validator.w3.org/nu/
  index=gen-apidocs/build/index.html

  sed -i 's#<BR />#<BR>#gi' $index
  sed -i 's#<BR/>#<BR>#gi' $index

  cp $index "/output/$release/"
done
