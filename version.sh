#!/bin/bash

cat ./internal/microservices/$1/swagger.yml | grep "version" | awk -F ' ' '{print $2}'
