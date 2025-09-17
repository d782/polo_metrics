package metrics

import (
	"context"
	"log"

	"github.com/d782/polo_metrics/internal/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetK8Metrics() ([]model.K8Metrics, error) {
	log.Println("Running K8 metrics ...")
	config, err := rest.InClusterConfig()

	if err != nil {
		log.Printf("error during configuration cluster: %v \n", err)
		return nil, err
	}

	if K8Client == nil {
		client, err := kubernetes.NewForConfig(config)

		if err != nil {
			log.Printf("error while creating K8 client %v \n", err)
			return nil, err
		}
		K8Client = client
	}

	pods, err := K8Client.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})

	if err != nil {
		log.Println("error while reading pods %v \n", err)
		return nil, err
	}
	var podsK8 []model.K8Metrics
	for _, pod := range pods.Items {
		k8Metrics := model.K8Metrics{
			Namespace:   pod.Namespace,
			Name:        pod.Name,
			StatusPhase: pod.Status.Phase,
			StartTime:   pod.Status.StartTime,
			PodIP:       pod.Status.PodIP,
			Labels:      pod.Labels,
		}

		var k8containerMetrics []model.K8ContainerMetrics

		for _, cs := range pod.Status.ContainerStatuses {

			cpu, _ := cs.AllocatedResources.Cpu().AsInt64()
			mem, _ := cs.AllocatedResources.Memory().AsInt64()

			k8container := model.K8ContainerMetrics{
				ContainerID:     cs.ContainerID,
				RestartCount:    cs.RestartCount,
				AllocatedMemory: mem,
				AllocatedCPU:    cpu,
				ImageID:         cs.ImageID,
			}

			if cs.State.Running != nil {
				k8container.State = "Running"
			} else if cs.State.Waiting != nil {
				k8container.State = "Waiting"
				k8container.Message = cs.State.Waiting.Message
				k8container.Reason = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil {
				k8container.State = "Terminated"
				k8container.Message = cs.State.Terminated.Message
				k8container.Reason = cs.State.Terminated.Reason
			}

			k8containerMetrics = append(k8containerMetrics, k8container)
		}
		k8Metrics.Containers = k8containerMetrics
		podsK8 = append(podsK8, k8Metrics)
	}

	return podsK8, nil
}
