package model

import (
	"time"

	"github.com/d782/polo_metrics/internal/metrics"
)

type JobProcess struct {
	Path     string
	Interval time.Duration
	Metric   metrics.IMetrics
}
