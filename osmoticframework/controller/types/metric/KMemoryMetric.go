package metric

type KMemoryPodMetric struct {
	Pod        string
	UsageBytes int64
}

type KMemoryNodeMetric struct {
	Node       string
	UsageBytes int64
}

type KMemoryPodPeakMetric struct {
	Pod        string
	UsageBytes int64
}

type KMemoryNodePeakMetric struct {
	Node       string
	UsageBytes int64
}

type KMemorySoftLimitMetric struct {
	Pod        string
	LimitBytes int64
}

type KMemoryPodReachLimitSecondsMetric struct {
	Pod                   string
	MemoryPressureSeconds int64
}
