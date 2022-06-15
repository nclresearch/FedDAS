package metric

type CpuEdgeMetric struct {
	Core  int
	Agent string
	Usage float64
}

type CpuContainerMetric struct {
	Container string
	Core      int
	Agent     string
	Usage     float64
}

type CpuEdgeTimeMetric struct {
	Core  int
	Agent string
	Time  float64
}

type CpuEdgeOverallUsageMetric struct {
	Agent string
	Usage float64
}
