package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/d782/polo_metrics/cmd/model"
)

func LoadEnv(prod bool) model.ConfigMetrics {

	log.Println("Loading environment variables ...")

	config := model.ConfigMetrics{
		IntervalHost:      5,
		IntervalProcess:   10,
		IntervalContainer: 15,
		IntervalK8:        20,
		EnableHost:        true,
		EnableProcess:     true,
		EnableContainer:   true,
		EnableK8:          true,
		ApiKey:            "default",
		UrlReports:        "http://localhost:8000",
	}
	if prod {
		config.IntervalHost = GetNumberEnv("INTERVAL_HOST")
		config.IntervalProcess = GetNumberEnv("INTERVAL_PROCESS")
		config.IntervalContainer = GetNumberEnv("INTERVAL_CONTAINER")
		config.EnableHost = GetBoolEnv("ENABLE_HOST")
		config.EnableProcess = GetBoolEnv("ENABLE_PROCESS")
		config.EnableContainer = GetBoolEnv("ENABLE_CONTAINER")
		config.ApiKey = os.Getenv("API_KEY")
		config.UrlReports = os.Getenv("URL_REPORTS")
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
