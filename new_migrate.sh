#!/bin/bash

microservice=$1
name=$2

path="./microservices/$microservice/migrate/"

migrate create --dir $path --name $name
