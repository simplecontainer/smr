#!/bin/bash
cd "$(dirname "$0")"
cd ..

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"

./smr apply definitions/example/configuration-mysql.yaml
./smr apply definitions/example/resource-nginx.yaml
./smr apply definitions/example/definition.yaml