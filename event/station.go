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

package event

type Station interface {
	Name() string
	IsRemote() bool
	IsBroadcast() bool
	Data() interface{}
}

type BaseStation struct {
	name    string
	usrData interface{}
}

type LocalStation struct {
	BaseStation
}

type RemoteStation struct {
	BaseStation
}

type BroadcastStation struct {
	BaseStation
}

func (bs *BaseStation) Name() string {
	return bs.name
}

func (bs *BaseStation) Data() interface{} {
	return bs.usrData
}

func NewLocalStation(name string, data interface{}) Station {
	return &LocalStation{
		BaseStation{
			name:    name,
			usrData: data,
		},
	}
}

func (*LocalStation) IsRemote() bool {
	return false
}

func (*LocalStation) IsBroadcast() bool {
	return false
}

func NewRemoteStation(name string, data interface{}) Station {
	return &RemoteStation{
		BaseStation{
			name:    name,
			usrData: data,
		},
	}
}

func (*RemoteStation) IsRemote() bool {
	return true
}

func (*RemoteStation) IsBroadcast() bool {
	return false
}

func NewBroadcastStation(name string, data interface{}) Station {
	return &BroadcastStation{
		BaseStation{
			name:    name,
			usrData: data,
		},
	}
}

func (*BroadcastStation) IsRemote() bool {
	return true
}

func (*BroadcastStation) IsBroadcast() bool {
	return true
}
