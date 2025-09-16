package model

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8Metrics struct {
	Namespace   string               `json:"namespace"`
	Name        string               `json:"name"`
	StatusPhase corev1.PodPhase      `json:"status_phase"`
	StartTime   *v1.Time             `json:"start_time"`
	PodIP       string               `json:"pod_ip,omitempty"`
	Labels      map[string]string    `json:"labels,omitempty"`
	Containers  []K8ContainerMetrics `json:"containers"`
}

type K8ContainerMetrics struct {
	ContainerID     string `json:"container_id"`
	Name            string `json:"name"`
	RestartCount    int32  `json:"restart_count"`
	AllocatedCPU    int64  `json:"allocated_cpu"`
	AllocatedMemory int64  `json:"allocated_memory"`
	State           string `json:"state"`
	Reason          string `json:"reason"`
	Message         string `json:"message"`
	ImageID         string `json:"image_id"`
}
