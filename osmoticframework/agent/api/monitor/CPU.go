package monitor

import (
	"fmt"
	"time"
)

//CPU related queries
//You will need to specify the agent ID (which refers to an edge device), container ID (If applicable), and time.

//Host average usage per core from the last 10 seconds (Separated by core ID)
func CPUEdgeAvg(time time.Time) (*Metric, error) {
	const query = "sum by (cpu) (rate(node_cpu_seconds_total{mode!='idle'}[10s]))"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Container average usage per core from the last 10 seconds (Separated by core ID)
func CPUContainerAvg(containerId string, time time.Time) (*Metric, error) {
	const query = "sum by (cpu, id) (rate(container_cpu_usage_seconds_total{id='/docker/%s'}[10s]))"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Cumulative CPU time by core
func CPUTimeTotal(time time.Time) (*Metric, error) {
	const query = "sum by (cpu) (node_cpu_seconds_total{mode!='idle'})"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Overall edge CPU utilization over the lastr 10 seconds
func CPUUtilization(time time.Time) (*Metric, error) {
	const query = "sum(sum by (cpu) (irate(node_cpu_seconds_total{mode!='idle'}[10s]))) / count(count(node_cpu_seconds_total) without (mode))"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
