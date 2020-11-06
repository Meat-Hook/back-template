#!/bin/bash

for service in ./internal/modules/*; do
  migrates=$service/init/migrate.sh
  chmod +x $migrates
  ./$migrates
done
