FROM golang:1.21.0-alpine

RUN apk add -U git bash make

COPY make.sh /

RUN \
  mkdir -p /go/src/github.com/kubernetes && \
  cd /go/src/github.com/kubernetes && \
  git clone https://github.com/xrstf/kubernetes-reference-docs reference-docs && \
  cd reference-docs && \
  git checkout additional-apis

WORKDIR /go/src/github.com/kubernetes/reference-docs
