FROM golang:1.21.0-alpine

RUN apk add -U git make

RUN \
  mkdir -p /go/src/github.com/kubernetes && \
  cd /go/src/github.com/kubernetes && \
  git clone https://github.com/kubernetes-sigs/reference-docs

WORKDIR /go/src/github.com/kubernetes/reference-docs
