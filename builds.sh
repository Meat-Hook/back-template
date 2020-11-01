#!/bin/sh

for service in ./internal/modules/*; do
  IFS='/' read -r -a path <<<"$service"
  name=${path[3]}
  echo "start service build - $name"
  rm -rf $service/bin
  mkdir $service/bin/
  go build -o $service/bin/ $service
done
