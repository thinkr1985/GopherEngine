package core

import (
	"log"
	"os"
	"runtime/pprof"
)

func StartCPUProfile() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
}

func StopCPUProfile() {
	pprof.StopCPUProfile()
}
