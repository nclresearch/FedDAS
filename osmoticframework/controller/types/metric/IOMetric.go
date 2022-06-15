package metric

type IOEdgeTimeMetric struct {
	Agent  string
	Device string
	Time   float64
}

type IOContainerTimeMetric struct {
	Container string
	Agent     string
	Device    string
	Time      float64
}

type IOEdgeBytesMetric struct {
	Agent  string
	Device string
	Bytes  uint64
}

type IOContainerBytesMetric struct {
	Container string
	Agent     string
	Device    string
	Bytes     uint64
}

type IOFilesystemBytesMetric struct {
	Agent      string
	Device     string
	MountPoint string
	Bytes      uint64
}
