package core

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
)

func GetCPU() string {
	info, err := cpu.Info()
	if err != nil {
		panic(err)
	}

	cpu := "UnKnownCPU"
	for _, ci := range info {
		cpu = ci.ModelName
		break
	}
	return cpu
}

// GetGPU returns the names of all video controllers (GPUs)
func GetGPU() string {
	cmd := exec.Command("wmic", "path", "win32_videocontroller", "get", "caption")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "Error fetching GPU info: " + err.Error()
	}

	lines := strings.Split(out.String(), "\n")

	var gpus []string
	for _, line := range lines[1:] { // Skip the header
		gpu := strings.TrimSpace(line)
		if gpu != "" {
			gpus = append(gpus, gpu)
		}
	}

	if len(gpus) == 0 {
		return "GPU not found"
	}
	return strings.Join(gpus, ", ")
}
