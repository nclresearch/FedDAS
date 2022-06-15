package monitor

import (
	"fmt"
	"time"
)

//Memory related queries
//You will need to specify the agent ID (which refers to an edge device), container ID (If applicable), and time.

//Memory usage in bytes in a container
func MemoryContainer(containerId string, time time.Time) (*Metric, error) {
	const query = "container_memory_usage_bytes{id='/docker/%s'}"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Memory usage in bytes in an edge device
func MemoryEdge(time time.Time) (*Metric, error) {
	const query = "node_memory_MemTotal_bytes - node_memory_MemFree_bytes - node_memory_Buffers_bytes - node_memory_Cached_bytes"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Highest memory usage in bytes over the last 5 minutes in a container
func MemoryContainerPeak(containerId string, time time.Time) (*Metric, error) {
	//container_memory_usage_bytes includes cached memory.
	const query = "max(max_over_time(container_memory_working_set_bytes{id='/docker/%s'}[5m]))"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Highest memory usage in bytes over the last 5 minutes in an edge device
func MemoryEdgePeak(time time.Time) (*Metric, error) {
	const query = "max_over_time(node_memory_MemTotal_bytes[5m]) - max_over_time(node_memory_MemFree_bytes[5m]) - max_over_time(node_memory_Buffers_bytes[5m]) - max_over_time(node_memory_Cached_bytes[5m])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//The time in seconds a container reaches its soft limit over the past 5 minutes
func MemoryContainerReachLimitSeconds(containerId string, time time.Time) (*Metric, error) {
	const query = "count_over_time((container_memory_working_set_bytes{id='/docker/%s'} > (container_spec_memory_reservation_limit_bytes{id='/docker/%s'} != 0))[5m:1s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
