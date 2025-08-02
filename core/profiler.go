package core

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/shirou/gopsutil/v3/mem"
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

func GetMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stat_text := fmt.Sprintf(
		"Alloc = %v MiB\n"+"\tTotalAlloc = %v MiB\n"+"\tSys = %v MiB\n"+"\tNumGC = %v",
		bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC,
	)
	return stat_text
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func GetMachineStats() string {
	// percent, _ := cpu.Percent(time.Second, false)
	vmStat, _ := mem.VirtualMemory()
	statsText := fmt.Sprintf(
		"Resolution: %dx%d\nMemory: %.1f/%.1f GB (%.1f%%)\n",

		SCREEN_WIDTH, SCREEN_HEIGHT,
		float64(vmStat.Used)/1e9, float64(vmStat.Total)/1e9, vmStat.UsedPercent,
	)
	return statsText
}
