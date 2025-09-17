package metrics

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/d782/polo_metrics/internal/model"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func GetContainerMetrics() ([]model.ContainerInfo, error) {
	log.Println("Running container metrics ...")
	var containerInfo []model.ContainerInfo

	if clientDocker == nil {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Println("error while starting docker client.")
			return nil, err
		}
		clientDocker = cli
	}

	containers, err := clientDocker.ContainerList(context.Background(), container.ListOptions{})

	if err != nil {
		log.Println("error while reading containers")
		return nil, err
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
	return containerInfo, nil
}
