package model

type ContainerInfo struct {
	Id       string           `json:"id"`
	Logs     string           `json:"logs"`
	CPU      uint64           `json:"cpu"`
	Mem      uint64           `json:"memory"`
	MemLimit uint64           `json:"memory_limit"`
	MemUsage uint64           `json:"memory_usage"`
	Network  NetworkContainer `json:"network"`
}

type NetworkContainer struct {
	RxBytes   uint64  `json:"rx_bytes"`
	RxPackets uint64  `json:"rx_packets"`
	TxBytes   uint64  `json:"tx_bytes"`
	TxPackets uint64  `json:"tx_packets"`
	Upload    float64 `json:"upload"`
	Download  float64 `json:"download"`
}
