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
	"io"
	"os"

	"github.com/ethereum/go-ethereum/log"
	colorable "github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

var glogger *log.GlogHandler

// LogConfig log config
type LogConfig struct {
	PrintOrigins bool   `mapstructure:"log-printorigins"`
	Level        int    `mapstructure:"log-level"`
	Vmodule      string `mapstructure:"log-vmodule"`
	BacktraceAt  string `mapstructure:"log-backtraceat"`
}

func defaultLogConfig() *LogConfig {
	return &LogConfig{
		PrintOrigins: false,
		Level:        3,
	}
}

func init() {
	usecolor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}

	glogger = log.NewGlogHandler(log.StreamHandler(output, log.TerminalFormat(usecolor)))
}

//Setup initializes logging based on the LogConfig
func (lc *LogConfig) Setup() {
	// logging
	log.PrintOrigins(lc.PrintOrigins)
	glogger.Verbosity(log.Lvl(lc.Level))
	glogger.Vmodule(lc.Vmodule)
	glogger.BacktraceAt(lc.BacktraceAt)
	log.Root().SetHandler(glogger)
}
