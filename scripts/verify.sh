#!/bin/bash

current_directory="$PWD"

cd $(dirname $0)
cd ../

go clean -testcache
TF_ACC=1 go test ./... -timeout 120m

result=$?

cd "$current_directory"

exit $result
