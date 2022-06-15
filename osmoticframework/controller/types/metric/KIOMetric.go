package metric

type KIONodeTimeMetric struct {
	Device string
	Time   float64
	Node   string
}

type KIOPodTimeMetric struct {
	Device string
	Time   float64
	Pod    string
}

type KIONodeReadMetric struct {
	Device    string
	BytesRead int64
	Node      string
}

type KIOPodReadMetric struct {
	Device    string
	BytesRead int64
	Pod       string
}

type KIONodeWriteMetric struct {
	Device       string
	BytesWritten int64
	Node         string
}

type KIOPodWriteMetric struct {
	Device       string
	BytesWritten int64
	Pod          string
}

type KIOFilesystemUsedMetric struct {
	Device     string
	Node       string
	Mountpoint string
	BytesUsed  int64
}

type KIOFilesystemSizeMetric struct {
	Device     string
	Node       string
	Mountpoint string
	BytesSize  int64
}
