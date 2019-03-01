#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
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

# Run the command inside the workspace.
cd "$ftdir/fractal"
PWD="$ftdir/fractal"
# Launch the arguments with the configured environment.

for arg in "$@"
do
   echo $arg
   $arg
done

