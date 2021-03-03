#!/bin/bash

module=$1
name=$2

path="./internal/modules/$module/migrate/"

migrate create --dir $path --name $name
