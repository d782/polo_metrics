package model

type SystemProcess struct {
	CPU    float64 `json:"cpu"`
	Memory float32 `json:"memory"`
	PID    int32   `json:"PID"`
	Name   string  `json:"process_name"`
}
