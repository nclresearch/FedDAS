package metric

type NetworkEdgeBytesMetric struct {
	Agent  string
	Bytes  uint64
	Device string
}

type NetworkContainerBytesMetric struct {
	Agent     string
	Container string
	Bytes     uint64
	Device    string
}

type NetworkEdgePacketsMetric struct {
	Agent   string
	Packets uint64
	Device  string
}

type NetworkContainerPacketsMetric struct {
	Agent     string
	Container string
	Packets   uint64
	Device    string
}

type NetworkEdgeErrorsMetric struct {
	Agent  string
	Errors uint64
	Device string
}

type NetworkContainerErrorsMetric struct {
	Agent     string
	Container string
	Errors    uint64
	Device    string
}
