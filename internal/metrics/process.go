package metrics

import (
	"log"

	"github.com/d782/polo_metrics/internal/model"
	"github.com/shirou/gopsutil/v3/process"
)

func ReadProcess() ([]model.SystemProcess, error) {
	log.Println("Running process metrics ...")
	process, err := process.Processes()

	var processList []model.SystemProcess

	if err != nil {
		return nil, err
	}

	for j := range process {
		cpu, err := process[j].CPUPercent()
		if err != nil {
			continue
		}
		mem, err := process[j].MemoryPercent()
		if err != nil {
			continue
		}

		name, err := process[j].Name()
		if err != nil {
			continue
		}
		if cpu < 75 && mem < 85 {
			continue
		}

		processLog := model.SystemProcess{
			CPU:    cpu,
			Memory: mem,
			PID:    process[j].Pid,
			Name:   name,
		}

		processList = append(processList, processLog)
	}

	return processList, nil
}
