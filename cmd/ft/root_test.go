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

package main

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func readAndUnmarshal(t *testing.T, file string) *ftConfig {
	v := viper.New()
	v.SetConfigFile(file)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("read config %v file err %v", file, err)
	}
	fc := new(ftConfig)
	err := v.Unmarshal(fc)
	if err != nil {
		t.Fatalf("unmarshal %v err %v", file, err)
	}
	return fc
}

func TestReadConfigFile(t *testing.T) {
	var (
		yamlFile = "./test/config.yaml"
		tomlFile = "./test/config.toml"
	)
	// test read config.yaml
	ftyamlcfg := readAndUnmarshal(t, yamlFile)
	// test read config.toml
	fttomlcfg := readAndUnmarshal(t, tomlFile)
	assert.Equal(t, ftyamlcfg, fttomlcfg)
}
