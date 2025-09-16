package metrics

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/d782/polo_metrics/internal/model"
	"github.com/d782/polo_metrics/internal/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientDocker *client.Client
var K8Client *kubernetes.Clientset
var prevNetStats model.NetMetrics
var prevContainerInfo []model.ContainerInfo
var prevCpuInfo []model.CpuMetrics

func GetMetricts() (*model.Metrics, error) {
	cpuPercent, err := cpu.Percent(time.Second, true)

	var cpuMetrics []model.CpuMetrics

	if err != nil {
		return nil, err
	}

	cpuOtherMetrics, err := cpu.Times(true)

	if err != nil {
		return nil, err
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
		return nil, err
	}

	memMetrics := model.MemMetrics{
		MemPercent: vm.UsedPercent,
		TotalMem:   float64(vm.Total),
	}

	disk, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	diskMetrics := model.DiskMetrics{
		DiskUsage: disk.UsedPercent,
		TotalDisk: float64(disk.Total),
	}

	netStats, err := net.IOCounters(false)

	if err != nil {
		return nil, err
	}

	if len(netStats) == 0 {
		return nil, err
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
	metrics := &model.Metrics{
		CPU:       cpuMetrics,
		MemoryRAM: memMetrics,
		DiskUsage: diskMetrics,
		Timestamp: time.Now().Unix(),
		Net:       netMetrics,
	}

	return metrics, nil
}

func ReadProcess() ([]model.SystemProcess, error) {
	log.Println("Reading process ...")
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

func GetContainerMetrics() []model.ContainerInfo {
	var containerInfo []model.ContainerInfo

	if clientDocker == nil {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Println("error while starting docker client.")
			return nil
		}
		clientDocker = cli
	}

	containers, err := clientDocker.ContainerList(context.Background(), container.ListOptions{})

	if err != nil {
		log.Println("error while reading containers")
		return nil
	}

	for _, c := range containers {
		statsResponse, err := clientDocker.ContainerStatsOneShot(context.Background(), c.ID)
		if err != nil {
			log.Println("error while reading container stats ...")
			continue
		}

		defer statsResponse.Body.Close()
		var stats container.StatsResponse
		if err := json.NewDecoder(statsResponse.Body).Decode(&stats); err != nil {
			log.Println("error parsing stats ...")
			continue
		}

		cpuDelta := stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
		systemDelta := stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
		currentCpus := stats.CPUStats.OnlineCPUs
		cpu := (cpuDelta / systemDelta) * uint64(currentCpus) * 100

		memUsage := stats.MemoryStats.Usage - stats.MemoryStats.Stats["cache"]
		memLimit := stats.MemoryStats.Limit
		memory := (memUsage / memLimit) * 100

		newContainerNetwork := model.NetworkContainer{
			RxBytes:   0,
			RxPackets: 0,
			TxBytes:   0,
			TxPackets: 0,
			Download:  0,
			Upload:    0,
		}

		for _, netStats := range stats.Networks {
			rxBytes := netStats.RxBytes
			rxPackets := netStats.RxPackets
			txBytes := netStats.RxBytes
			txPackets := netStats.TxPackets

			newContainerNetwork.RxBytes = rxBytes
			newContainerNetwork.RxPackets = rxPackets
			newContainerNetwork.TxBytes = txBytes
			newContainerNetwork.TxPackets = txPackets

			if len(prevContainerInfo) > 0 {
				for _, prevContainer := range prevContainerInfo {
					if prevContainer.Id == c.ID {
						RxBytesSec := (newContainerNetwork.RxBytes - prevContainer.Network.RxBytes) / 5
						TxBytesSec := (newContainerNetwork.TxBytes - prevContainer.Network.TxBytes) / 5
						newContainerNetwork.Download = (float64(RxBytesSec) / 1024)
						newContainerNetwork.Upload = (float64(TxBytesSec) / 1024)
					}
				}
			}
		}

		logs, err := clientDocker.ContainerLogs(context.Background(), c.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "20",
		})

		if err != nil {
			log.Println("error reading logs")
			continue
		}

		defer logs.Close()

		logBytes, _ := io.ReadAll(logs)

		logsParsed := string(logBytes)

		newContainerInfo := model.ContainerInfo{
			CPU:     cpu,
			Mem:     memory,
			Id:      c.ID,
			Logs:    logsParsed,
			Network: newContainerNetwork,
		}

		containerInfo = append(containerInfo, newContainerInfo)
	}
	prevContainerInfo = containerInfo
	return containerInfo
}

func GetK8Metrics() []model.K8Metrics {
	config, err := rest.InClusterConfig()

	if err != nil {
		log.Printf("error during configuration cluster: %v \n", err)
		return nil
	}

	if K8Client == nil {
		client, err := kubernetes.NewForConfig(config)

		if err != nil {
			log.Printf("error while creating K8 client %v \n", err)
			return nil
		}
		K8Client = client
	}

	pods, err := K8Client.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})

	if err != nil {
		log.Println("error while reading pods %v \n", err)
		return nil
	}
	var podsK8 []model.K8Metrics
	for _, pod := range pods.Items {
		k8Metrics := model.K8Metrics{
			Namespace:   pod.Namespace,
			Name:        pod.Name,
			StatusPhase: pod.Status.Phase,
			StartTime:   pod.Status.StartTime,
			PodIP:       pod.Status.PodIP,
			Labels:      pod.Labels,
		}

		var k8containerMetrics []model.K8ContainerMetrics

		for _, cs := range pod.Status.ContainerStatuses {

			cpu, _ := cs.AllocatedResources.Cpu().AsInt64()
			mem, _ := cs.AllocatedResources.Memory().AsInt64()

			k8container := model.K8ContainerMetrics{
				ContainerID:     cs.ContainerID,
				RestartCount:    cs.RestartCount,
				AllocatedMemory: mem,
				AllocatedCPU:    cpu,
				ImageID:         cs.ImageID,
			}

			if cs.State.Running != nil {
				k8container.State = "Running"
			} else if cs.State.Waiting != nil {
				k8container.State = "Waiting"
				k8container.Message = cs.State.Waiting.Message
				k8container.Reason = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil {
				k8container.State = "Terminated"
				k8container.Message = cs.State.Terminated.Message
				k8container.Reason = cs.State.Terminated.Reason
			}

			k8containerMetrics = append(k8containerMetrics, k8container)
		}
		k8Metrics.Containers = k8containerMetrics
		podsK8 = append(podsK8, k8Metrics)
	}

	return podsK8
}
