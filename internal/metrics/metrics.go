package metrics

import (
	"log"
	"time"

	"github.com/d782/polo_metrics/internal/model"
	"github.com/d782/polo_metrics/internal/utils"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var prevNetStats model.NetMetrics

var prevCpuInfo []model.CpuMetrics

func GetMetricts() (model.Metrics, error) {
	var cpuMetrics []model.CpuMetrics

	log.Println("Running system metrics ...")
	cpuPercent, err := cpu.Percent(time.Second, true)

	if err != nil {
		return model.Metrics{}, err
	}

	info, err := host.Info()
	if err != nil {
		return model.Metrics{}, err
	}

	cpuOtherMetrics, err := cpu.Times(true)

	if err != nil {
		return model.Metrics{}, err
	}

	for i, cpu := range cpuPercent {
		newCpuMetric := model.CpuMetrics{}
		var userPercent float64
		var systemPercent float64
		var idlePercent float64
		if len(prevCpuInfo) > 0 {
			dSystem := cpuOtherMetrics[i].System - prevCpuInfo[i].System
			dUser := cpuOtherMetrics[i].User - prevCpuInfo[i].User
			dIdle := cpuOtherMetrics[i].Idle - prevCpuInfo[i].Idle
			total := dUser + dSystem + dIdle
			userPercent = (dUser / total) * 100
			systemPercent = (dSystem / total) * 100
			idlePercent = (dIdle / total) * 100
		} else {
			total := cpuOtherMetrics[i].System + cpuOtherMetrics[i].Idle + cpuOtherMetrics[i].User
			userPercent = (cpuOtherMetrics[i].User / total) * 100
			systemPercent = (cpuOtherMetrics[i].System / total) * 100
			idlePercent = (cpuOtherMetrics[i].Idle / total) * 100
		}
		newCpuMetric.CPUPercent = cpu
		newCpuMetric.System = utils.Clamp(systemPercent)
		newCpuMetric.Idle = utils.Clamp(idlePercent)
		newCpuMetric.User = utils.Clamp(userPercent)

		cpuMetrics = append(cpuMetrics, newCpuMetric)
	}

	prevCpuInfo = cpuMetrics

	vm, err := mem.VirtualMemory()

	if err != nil {
		return model.Metrics{}, err
	}

	memMetrics := model.MemMetrics{
		MemPercent: vm.UsedPercent,
		TotalMem:   float64(vm.Total),
	}

	disk, err := disk.Usage("/")
	if err != nil {
		return model.Metrics{}, err
	}

	diskMetrics := model.DiskMetrics{
		DiskUsage: disk.UsedPercent,
		TotalDisk: float64(disk.Total),
	}

	netStats, err := net.IOCounters(false)

	if err != nil {
		return model.Metrics{}, err
	}

	if len(netStats) == 0 {
		return model.Metrics{}, err
	}

	netMetrics := model.NetMetrics{
		BytesSent: netStats[0].BytesSent,
		BytesRcv:  netStats[0].BytesRecv,
		Upload:    0,
		Download:  0,
	}

	if prevNetStats.BytesSent != 0 {
		sentPerSec := (netMetrics.BytesSent - prevNetStats.BytesSent) / 5
		recvPerSec := (netMetrics.BytesRcv - prevNetStats.BytesRcv) / 5

		netMetrics.Upload = float64(sentPerSec / (1024 * 1024))
		netMetrics.Download = float64(recvPerSec / (1024 * 1024))
	}
	prevNetStats = netMetrics

	metrics := model.Metrics{
		Name:          info.Hostname,
		Platform:      info.Platform,
		OS:            info.OS,
		KernelVersion: info.KernelVersion,
		CPU:           cpuMetrics,
		MemoryRAM:     memMetrics,
		DiskUsage:     diskMetrics,
		Timestamp:     time.Now().Unix(),
		Net:           netMetrics,
	}

	return metrics, nil
}
