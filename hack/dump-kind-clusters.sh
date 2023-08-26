#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

mkdir -p data

touch kubeconfig
export KUBECONFIG="$(realpath kubeconfig)"

clusterName=kubernetes-apis

for release in 1.21 1.22 1.23 1.24 1.25 1.26 1.27; do
  echo "Dumping APIs for Kubernetes $release …"

  image="$(jq --arg rel "$release" -r '.[$rel]' hack/kind-images.json)"

  kind create cluster \
    --image "$image" \
    --name "$clusterName"

  _build/dumper > "data/release-$release.json"

  kind delete cluster --name "$clusterName"
done
