package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/d782/polo_metrics/cmd/model"
	"github.com/d782/polo_metrics/internal/api"
	"github.com/d782/polo_metrics/internal/metrics"
)

func RunJob(name string, intervalTime time.Duration, job func()) {
	ticker := time.NewTicker(intervalTime)

	defer ticker.Stop()

	for range ticker.C {
		log.Println("Executing ... ", name)
		job()
	}
}

func LoadEnv() model.ConfigMetrics {
	intervalHost := GetNumberEnv("INTERVAL_HOST")
	intervalProcess := GetNumberEnv("INTERVAL_PROCESS")
	intervalContainer := GetNumberEnv("INTERVAL_CONTAINER")
	enableHost := GetBoolEnv("ENABLE_HOST")
	enableProcess := GetBoolEnv("ENABLE_PROCESS")
	enableContainer := GetBoolEnv("ENABLE_CONTAINER")
	apiKey := os.Getenv("API_KEY")
	url := os.Getenv("URL_REPORTS")

	config := model.ConfigMetrics{
		IntervalHost:      intervalHost,
		IntervalProcess:   intervalProcess,
		IntervalContainer: intervalContainer,
		EnableHost:        enableHost,
		EnableProcess:     enableProcess,
		EnableContainer:   enableContainer,
		ApiKey:            apiKey,
		UrlReports:        url,
	}

	return config
}

func GetNumberEnv(key string) int64 {
	variable := os.Getenv(key)
	result, err := strconv.Atoi(variable)
	if err != nil {
		return 5
	}
	return int64(result)
}

func GetBoolEnv(key string) bool {
	variable := os.Getenv(key)
	result, err := strconv.ParseBool(variable)
	if err != nil {
		return false
	}
	return result
}

func main() {
	log.Println("Polo metrics agent started")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go RunJob("system", 10*time.Second, func() {
		m, err := metrics.GetMetricts()
		if err != nil {
			log.Println("Unable to read system metrics")
			log.Println("... trying again")
			time.Sleep(2 * time.Second)
			return
		}

		go api.SendReport("http://localhost:8000", "system", m, 10*time.Second)
	})

	go RunJob("process", 15*time.Second, func() {
		processList, err := metrics.ReadProcess()

		if err != nil {
			log.Println("Unable to read process")
			log.Printf("... trying again")
			time.Sleep(2 * time.Second)
			return
		}
		if len(processList) > 0 {
			api.SendReport("localhost:8000", "process", processList, 15*time.Second)
		}
	})

	go RunJob("containers", 20*time.Second, func() {
		containersInfo := metrics.GetContainerMetrics()
		if len(containersInfo) > 0 {
			api.SendReport("localhost:8000", "containers", containersInfo, 20*time.Second)
		}
	})

	go RunJob("k8", 25*time.Second, func() {
		k8Metrics := metrics.GetK8Metrics()
		if len(k8Metrics) > 0 {
			api.SendReport("localhost:8000", "k8", k8Metrics, 25*time.Second)
		}
	})

	select {}
}
