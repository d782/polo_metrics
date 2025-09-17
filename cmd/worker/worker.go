package worker

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/d782/polo_metrics/internal/api"
	"github.com/d782/polo_metrics/internal/metrics"
)

func RunJob(ctx context.Context, name string, intervalTime time.Duration, job func()) {
	ticker := time.NewTicker(intervalTime)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping service worker :", name)
			return
		case <-ticker.C:
			log.Println("Executing ... ", name)
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("System recovered from error during job execution:", r)
					}
				}()
				job()
			}()

		}
	}
}

func RunMetricWorker(ctx context.Context, url string, path string, interval time.Duration, strategy metrics.IMetrics) {
	RunJob(ctx, path, interval, func() {
		data, err := strategy.Collect()
		if err != nil {
			log.Println("Failed to executed operation ...")
		}

		if data != nil {
			val := reflect.ValueOf(data)
			if (val.Kind() == reflect.Slice || val.Kind() == reflect.Array) && val.Len() == 0 {
				return
			}
			go api.SendReport(url, path, data, interval)
		}
	})
}
