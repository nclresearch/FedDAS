package monitor

import (
	"fmt"
	"time"
)

//Disk IO related queries
//You will need to specify the agent ID (which refers to an edge device), container ID (If applicable), and time.

//Number of seconds spent doing IO operations on an edge device over the last 1 minute
//If the time is high, either the system is thrashing or there's a hard drive failure.
func IOEdgeTime(time time.Time) (*Metric, error) {
	const query = "rate(node_disk_io_time_seconds_total[1m])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Number of seconds spent doing IO operations on a container over the last 1 minute
func IOContainerTime(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_fs_io_time_seconds{id='/docker/%s'}[1m])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Bytes read on an edge device over the last 10 seconds
func IOReadEdgeBytes(time time.Time) (*Metric, error) {
	const query = "rate(node_disk_read_bytes_total[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Bytes read on a container over the last 10 seconds
func IOReadContainerBytes(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_fs_reads_bytes_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Bytes written on an edge device over the last 10 seconds
func IOWriteEdgeBytes(time time.Time) (*Metric, error) {
	const query = "rate(node_disk_written_bytes_total[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Bytes written on a container over the last 10 seconds
func IOWriteContainerBytes(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_fs_writes_bytes_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Total disk space used in bytes. Edge side only
func IOFilesystemUsedBytes(time time.Time) (*Metric, error) {
	const query = "node_filesystem_size_bytes - node_filesystem_avail_bytes"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Total disk size in bytes. Edge size only
func IOFilesystemSizeBytes(time time.Time) (*Metric, error) {
	const query = "node_filesystem_size_bytes"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
