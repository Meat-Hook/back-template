#!/bin/bash

module=$1
name=$2

path="./internal/modules/$module/internal/repo/migrate"

migrate create --dir $path --name $name
