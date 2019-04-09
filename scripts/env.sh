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


#!/bin/sh

set -e

if [ ! -f "scripts/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
ftdir="$workspace/src/github.com/fractalplatform"
bindir="$PWD/build/bin"
mkdir -p $bindir
if [ ! -L "$ftdir/fractal" ]; then
    mkdir -p "$ftdir"
    cd "$ftdir"
    ln -s ../../../../../. fractal
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

pwd

# Launch the arguments with the configured environment.
for arg in $@
do
    echo $arg
    $arg
done

