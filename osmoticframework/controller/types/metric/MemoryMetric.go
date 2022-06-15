package metric

type MemoryContainerMetric struct {
	Container string
	Agent     string
	Usage     uint64
}

type MemoryEdgeMetric struct {
	Agent string
	Usage uint64
}

type MemoryContainerLimitSecondsMetric struct {
	Container string
	Agent     string
	Time      uint64
}
