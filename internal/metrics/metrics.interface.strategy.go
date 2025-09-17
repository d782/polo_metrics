package metrics

type IMetrics interface {
	Collect() (interface{}, error)
}

type SystemMetrics struct{}

func (s SystemMetrics) Collect() (interface{}, error) {
	return GetMetricts()
}

type ProcessMetrics struct{}

func (p ProcessMetrics) Collect() (interface{}, error) {
	return ReadProcess()
}

type ContainerMetrics struct{}

func (c ContainerMetrics) Collect() (interface{}, error) {
	return GetContainerMetrics()
}

type KubernetesMetrics struct{}

func (k KubernetesMetrics) Collect() (interface{}, error) {
	return GetK8Metrics()
}
