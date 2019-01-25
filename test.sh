#!/usr/bin/env bash

set -e
echo "mode: count" >coverage.out

for d in $(go list ./... | grep -v vendor | grep -v test); do
    go test -coverprofile=profile.out -covermode=count $d
    if [ -f profile.out ]; then
        cat profile.out | grep -v "mode: count" | cat >> coverage.out
        rm profile.out
    fi
done
