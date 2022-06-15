package metric

type KCPUUsage struct {
	Core  int
	Node  string
	Usage float64
}

type KCPUTime struct {
	Core    int
	Node    string
	CPUTime float64
}

type KCPUPodUsage struct {
	Pod   string
	Usage float64
}

type KCPUTotalUsage struct {
	Node  string
	Usage float64
}
