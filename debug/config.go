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

package debug

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fjl/memsize/memsizeui"
)

var Memsize memsizeui.Handler

type Config struct {
	Pprof            bool   `mapstructure:"pprof"`
	PprofPort        int    `mapstructure:"pprofport"`
	PprofAddr        string `mapstructure:"pprofaddr"`
	Memprofilerate   int    `mapstructure:"memprofilerate"`
	Blockprofilerate int    `mapstructure:"blockprofilerate"`
	Cpuprofile       string `mapstructure:"cpuprofile"`
	Trace            string `mapstructure:"trace"`
}

func DefaultConfig() *Config {
	return &Config{
		Pprof:          false,
		PprofPort:      6060,
		PprofAddr:      "localhost",
		Memprofilerate: runtime.MemProfileRate,
	}
}

// Setup initializes profiling ,It should be called
// as early as possible in the program.
func Setup(debugCfg *Config) error {
	// profiling, tracing
	runtime.MemProfileRate = debugCfg.Memprofilerate
	Handler.SetBlockProfileRate(debugCfg.Blockprofilerate)
	if debugCfg.Trace != "" {
		if err := Handler.StartGoTrace(debugCfg.Trace); err != nil {
			return err
		}
	}
	if debugCfg.Cpuprofile != "" {
		if err := Handler.StartCPUProfile(debugCfg.Cpuprofile); err != nil {
			return err
		}
	}

	// pprof server
	if debugCfg.Pprof {
		address := fmt.Sprintf("%s:%d", debugCfg.PprofAddr, debugCfg.PprofPort)
		StartPProf(address)
	}
	return nil
}

func StartPProf(address string) {
	http.Handle("/memsize/", http.StripPrefix("/memsize", &Memsize))
	log.Info("Starting pprof server", "addr", fmt.Sprintf("http://%s/debug/pprof", address))
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Error("Failure in running pprof server", "err", err)
		}
	}()
}

// Exit stops all running profiles, flushing their output to the
// respective file.
func Exit() {
	Handler.StopCPUProfile()
	Handler.StopGoTrace()
}
