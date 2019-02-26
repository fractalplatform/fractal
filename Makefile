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


all: ft ftkey ftfinder

ft:
	@./build/env.sh \
	"go install ./cmd/ft" \
	"go build ./cmd/ft" \
	"mv ft ./build/bin" \

ftkey:
	@./build/env.sh \
	"go install ./cmd/ftkey" \
	"go build ./cmd/ftkey" \
	"mv ftkey ./build/bin"

ftfinder:
	@./build/env.sh \
	"go install ./cmd/ftfinder" \
	"go build ./cmd/ftfinder" \
	"mv ftfinder ./build/bin"

clean:
	rm -rf build/bin build/_workspace

test:
	./build/env.sh ./test.sh

.PHONY: all clean test ftfinder ft ftkey ftfinder
