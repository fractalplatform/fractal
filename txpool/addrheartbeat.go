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

package txpool

import (
	"time"

	"github.com/fractalplatform/fractal/common"
)

// nameByHeartbeat is an account name tagged with its last activity timestamp.
type nameByHeartbeat struct {
	name      common.Name
	heartbeat time.Time
}

type namesByHeartbeat []nameByHeartbeat

func (n namesByHeartbeat) Len() int           { return len(n) }
func (n namesByHeartbeat) Less(i, j int) bool { return n[i].heartbeat.Before(n[j].heartbeat) }
func (n namesByHeartbeat) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
