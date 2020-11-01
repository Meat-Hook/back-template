#!/bin/sh

cat internal/modules/$1/swagger.yml | grep "version" | awk -F ' ' '{print $2}'
