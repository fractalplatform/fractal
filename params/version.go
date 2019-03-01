// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package params

import "fmt"

const (
	//ClientIdentifier Client identifier to advertise over the network
	ClientIdentifier = "ft"
)

const (
	// VersionMajor is Major version component of the current release
	VersionMajor = 0
	// VersionMinor is Minor version component of the current release
	VersionMinor = 0
	// VersionPatch is Patch version component of the current release
	VersionPatch = 4
	// VersionMeta is Version metadata to append to the version string
	VersionMeta = "unstable"
)

// Version holds the textual version string.
var Version = func() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}()

// ArchiveVersion holds the textual version string used for Geth archives.
func ArchiveVersion(gitCommit string) string {
	vsn := Version
	if VersionMeta != "stable" {
		vsn += "-" + VersionMeta
	}
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	return vsn
}
