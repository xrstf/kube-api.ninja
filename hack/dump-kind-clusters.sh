#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

mkdir -p data

touch kubeconfig
export KUBECONFIG="$(realpath kubeconfig)"

clusterName=kubernetes-apis

downloadKind() {
  local kubernetesRelease="$1"
  local arch="$(go env GOARCH)"

  local kindVersion
  kindVersion="$(jq --arg rel "$release" -r '.[$rel].kind' hack/kind-configs.json)"

  local filename="_build/kind-$kindVersion"
  if [ ! -f "$filename" ]; then
    wget --output-document "$filename" "https://github.com/kubernetes-sigs/kind/releases/download/v$kindVersion/kind-linux-$arch"
    chmod +x "$filename"
  fi

  echo "$filename"
}

for release in 1.11 1.12 1.13 1.14 1.15 1.16 1.17 1.18 1.19 1.20 1.21 1.22 1.23 1.24 1.25 1.26 1.27 1.28; do
  echo "Dumping APIs for Kubernetes $release …"

  kind="$(downloadKind "$release")"

  image="$(jq --arg rel "$release" -r '.[$rel].image' hack/kind-configs.json)"

  $kind create cluster \
    --image "$image" \
    --name "$clusterName"

  _build/dumper > "data/release-$release.json"

  $kind delete cluster --name "$clusterName"
done
