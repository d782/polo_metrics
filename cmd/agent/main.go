package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/d782/polo_metrics/cmd/model"
	"github.com/d782/polo_metrics/cmd/utils"
	"github.com/d782/polo_metrics/cmd/worker"
	"github.com/d782/polo_metrics/internal/metrics"
)

func main() {
	var wg sync.WaitGroup

	log.Println("Starting polo metrics agent ...")
	ctx, cancel := context.WithCancel(context.Background())

	systemSignal := make(chan os.Signal, 1)
	signal.Notify(systemSignal, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var jobs []model.JobProcess
	env := utils.LoadEnv(false)
	jobs = GetJobs(jobs, env)

	for _, j := range jobs {
		wg.Add(1)
		go func(job model.JobProcess) {
			defer wg.Done() // <- aquí
			worker.RunMetricWorker(ctx, env.UrlReports, job.Path, job.Interval, job.Metric)
		}(j)
	}

	<-systemSignal
	log.Println("Stopping metrics agent ...")
	cancel()
	wg.Wait()
	log.Println("All workers stopped.")
	time.Sleep(time.Second)
}

func GetJobs(j []model.JobProcess, env model.ConfigMetrics) []model.JobProcess {

	if env.EnableHost {
		j = append(j, model.JobProcess{
			Path:     "system",
			Interval: time.Duration(env.IntervalHost) * time.Second,
			Metric:   metrics.SystemMetrics{},
		})
	}

	if env.EnableProcess {
		j = append(j, model.JobProcess{
			Path:     "process",
			Interval: time.Duration(env.IntervalProcess) * time.Second,
			Metric:   metrics.ProcessMetrics{},
		})
	}

	if env.EnableContainer {
		j = append(j, model.JobProcess{
			Path:     "containers",
			Interval: time.Duration(env.IntervalContainer) * time.Second,
			Metric:   metrics.ContainerMetrics{},
		})
	}

	if env.EnableK8 {
		j = append(j, model.JobProcess{
			Path:     "k8",
			Interval: time.Duration(env.IntervalK8) * time.Second,
			Metric:   metrics.KubernetesMetrics{},
		})
	}

	return j
}
