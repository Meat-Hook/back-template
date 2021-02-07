#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

for service in ./internal/modules/*; do
  IFS='/' read -r -a path <<<"$service"
  name=${path[3]}
  echo "build service - $name"
  rm -rf $service/bin
  mkdir $service/bin/
  go build -o $service/bin/ $service
done
