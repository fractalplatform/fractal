# Copyright 2018 The Fractal Team Authors
# This file is part of the fractal project.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

#!/usr/bin/env bash

# clear test_sdk data
rm -r ./build/test_sdk

# generate test directory
mkdir ./build/test_sdk

# clear test node 
ps -ef | grep ./build/test_sdk/ft | grep -v grep |  awk -F ' ' '{print $2}' | xargs kill -9

# start test node
./build/bin/ft --datadir ./build/test_sdk/ft --log_level=4 --miner_start > ./build/test_sdk/test.log 2>&1 &

# collect code coverrage data
set -e
echo "mode: count" >coverage.out

for d in $(go list ./... | grep -v vendor | grep -v test); do
    echo testing $d ...
    go test -coverprofile=profile.out -covermode=count $d
    if [ -f profile.out ]; then
        cat profile.out | grep -v "mode: count" | cat >> coverage.out
        rm profile.out
    fi
done

# kill test node 
ps -ef | grep ./build/test_sdk/ft | grep -v grep |  awk -F ' ' '{print $2}' | xargs kill -9

# clear test_sdk data
rm -r ./build/test_sdk

