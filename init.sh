#!/bin/bash

for service in ./internal/modules/*; do
  initFile=$service/init/init.sh
  chmod +x $initFile
  IFS='/' read -r -a path <<<"$service"
  name=${path[3]}
  echo "init service $name"
  ./$initFile
done
