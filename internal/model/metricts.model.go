package model

type Metrics struct {
	CPU       []CpuMetrics `json:"cpu"`
	MemoryRAM MemMetrics   `json:"memory_ram"`
	DiskUsage DiskMetrics  `json:"disk"`
	Timestamp int64        `json:"timestamp"`
	Net       NetMetrics   `json:"net_metrics"`
}

type NetMetrics struct {
	BytesSent uint64  `json:"bytes_sent"`
	BytesRcv  uint64  `json:"bytes_recv"`
	Upload    float64 `json:"upload"`
	Download  float64 `json:"download"`
}

type CpuMetrics struct {
	CPUPercent float64 `json:"cpu_percent"`
	User       float64 `json:"user"`
	System     float64 `json:"system"`
	Idle       float64 `json:"idle"`
}

type MemMetrics struct {
	MemPercent float64 `json:"memory_parcent"`
	TotalMem   float64 `json:"total_memory"`
}

type DiskMetrics struct {
	DiskUsage float64 `json:"disk_usage"`
	TotalDisk float64 `json:"total_disk"`
}
