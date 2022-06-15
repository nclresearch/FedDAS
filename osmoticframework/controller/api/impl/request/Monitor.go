package request

import (
	"encoding/json"
	"github.com/lithammer/shortuuid"
	"github.com/streadway/amqp"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
	"time"
)

/*
All edge monitoring API functions
Prometheus smallest scale in time is second. Therefore there is no point in using nanosecond precision.
*/

func CPUEdgeAvgRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "cpu_edge_avg"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func CPUContainerAvgRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "cpu_container_avg"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func CPUTimeRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "cpu_time"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func CPUUtilizationRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "cpu_utilization"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func MemoryEdgeRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "memory_edge"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func MemoryContainerRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "memory_container"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func MemoryEdgePeakRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "memory_edge_peak"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func MemoryContainerPeakRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "memory_container_peak"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func MemoryContainerLimitSecondsRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "memory_container_limit_seconds"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func IOEdgeTimeRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_edge_time"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func IOContainerTimeRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_container_time"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func IOEdgeReadRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_edge_read"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func IOContainerReadRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_container_read"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func IOEdgeWriteRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_edge_write"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func IOContainerWriteRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_container_write"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func IOFilesystemUsedRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_filesystem_used"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func IOFilesystemSizeRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "io_filesystem_size"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeRxBytesRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_rx_bytes"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeRxPacketsRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_rx_packets"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeRxDroppedRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_rx_dropped"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeRxErrorRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_rx_error"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeTxBytesRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_tx_bytes"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeTxPacketsRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_tx_packets"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeTxDroppedRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_tx_dropped"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetEdgeTxErrorRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_edge_tx_error"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func NetContainerRxBytesRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_rx_bytes"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerRxPacketsRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_rx_packets"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerRxDroppedRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_rx_dropped"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerRxErrorRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_rx_error"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerTxBytesRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_tx_bytes"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerTxPacketsRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_tx_packets"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerTxDroppedRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_tx_dropped"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func NetContainerTxErrorRequest(agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "net_container_tx_error"
	return containerMonitorRequest(command, agentId, containerId, timestamp, timeout)
}

func ThermalRequest(agentId string, timestamp time.Time, timeout float64) *RequestTask {
	const command = "thermal"
	return edgeMonitorRequest(command, agentId, timestamp, timeout)
}

func containerMonitorRequest(command, agentId, containerId string, timestamp time.Time, timeout float64) *RequestTask {
	var id string
	for true {
		id := shortuuid.New()
		if _, ok := MonitorRequests.Load(id); !ok {
			break
		}
	}
	request, err := json.Marshal(Request{
		RequestID: id,
		Command:   command,
		Args: map[string]interface{}{
			"edge":        agentId,
			"time":        timestamp.Unix(),
			"containerId": containerId,
		},
	})
	if err != nil {
		log.Error.Println("Failed constructing monitor API command")
		log.Error.Println(err)
		return nil
	}
	err = queue.Ch.Publish(
		"",
		"monitor-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending monitor API request")
		log.Error.Println(err)
		return nil
	}
	MonitorRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "monitor",
		Command: command,
		Ack:     false,
		Time:    time.Now(),
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "monitor",
		Command: command,
		Time:    time.Now(),
		Args: map[string]interface{}{
			"edge":        agentId,
			"time":        timestamp.Unix(),
			"containerId": containerId,
		},
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	MonitorTaskList.Store(id, task)
	return &task
}

func edgeMonitorRequest(command, agentId string, timestamp time.Time, timeout float64) *RequestTask {
	var id string
	for true {
		id := shortuuid.New()
		if _, ok := MonitorRequests.Load(id); !ok {
			break
		}
	}
	log.Info.Printf("%s << Monitoring request %s\n", agentId, command)
	request, err := json.Marshal(Request{
		RequestID: id,
		Command:   command,
		Args: map[string]interface{}{
			"edge": agentId,
			"time": timestamp.Unix(),
		},
	})
	if err != nil {
		log.Error.Println("Failed constructing monitor API command")
		log.Error.Println(err)
		return nil
	}
	err = queue.Ch.Publish(
		"",
		"monitor-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending monitor API request")
		log.Error.Println(err)
		return nil
	}
	MonitorRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "monitor",
		Command: command,
		Ack:     false,
		Time:    time.Now(),
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "monitor",
		Command: command,
		Time:    time.Now(),
		Args: map[string]interface{}{
			"edge": agentId,
			"time": timestamp.Unix(),
		},
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	MonitorTaskList.Store(id, task)
	return &task
}
