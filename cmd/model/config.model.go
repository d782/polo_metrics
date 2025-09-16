package model

type ConfigMetrics struct {
	IntervalHost      int64
	IntervalProcess   int64
	IntervalContainer int64
	EnableHost        bool
	EnableProcess     bool
	EnableContainer   bool
	ApiKey            string
	UrlReports        string
}
