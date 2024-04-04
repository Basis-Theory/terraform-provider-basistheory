#!/bin/bash

current_directory="$PWD"

cd $(dirname $0)
cd ../

go generate

result=$?

cd "$current_directory"

exit $result
