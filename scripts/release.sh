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

version_regex="^v[0-9]+\.[0-9]+\.[0-9]+$"

set -e

function release {
    notes="NOTES.md"
    echo "Building and releasing $tag..."
    echo "Building and pushing binaries"
    [[ -e "$notes" ]] && goreleaser --release-notes "$notes" --rm-dist
}

# If passed argument try to use that as tag otherwise read from local repo
if [[ $1 ]]; then
    # Override mode, try to release this tag
    export tag=$1
else
    echo "Getting tag from HEAD which is $(git rev-parse HEAD)"
    export tag=$(git tag --points-at HEAD)
fi

if [[ ! ${tag} ]]; then
    echo "No tag so not releasing."
    exit 0
fi

# Only release semantic version syntax tags
if [[ ! ${tag} =~ ${version_regex} ]] ; then
    echo "Tag '$tag' does not match version regex '$version_regex' so not releasing."
    exit 0
fi

release
