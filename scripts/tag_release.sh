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

set -e

# Wait a second so we don't see ephemeral file changes
sleep 1

# Don't tag if there is a dirty working dir
if ! git diff-index --quiet HEAD  ; then
    echo "Warning there appears to be uncommitted changes in the working directory:"
    git diff-index HEAD
    echo
    echo "Please commit them or stash them before tagging a release."
    echo
fi

version=v$(go run ./cmd/project/main.go version )
notes=$(go run ./cmd/project/main.go notes)

echo "This command will tag the current commit $(git rev-parse --short HEAD) as version $version"
echo
echo "$notes" | sed 's/^/> /'
echo
echo "It will then push the version tag to origin."
echo
read -p "Do you want to continue? [Y\n]: " -r
# Just hitting return defaults to continuing
[[ $REPLY ]] && [[ ! $REPLY =~ ^[Yy]$ ]] && echo && exit 0
echo


# Create tag
echo "Tagging version $version with message:"
echo ""
echo "$notes"
echo ""
echo "$notes" | git tag -a ${version} -F-

# Push tag
git push origin ${version}