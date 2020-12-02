#!/bin/bash

for service in ./internal/modules/*; do
  migrates=$service/migrate.sh
  chmod +x $migrates
  ./$migrates
done