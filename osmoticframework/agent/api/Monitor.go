package api

import (
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	monitor2 "osmoticframework/agent/api/monitor"
	"osmoticframework/agent/log"
	"time"
)

//Endpoints of the monitoring API
//It only handles the arguments of the requests. For the actual request processing, see the monitor directory

//Host CPU usage per core from the last 10 seconds (Separated by core ID)
func CPUEdgeAvgEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.CPUEdgeAvg(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func CPUContainerAvgEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.CPUContainerAvg(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func CPUTimeEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.CPUTimeTotal(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func CPUUtilizeEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.CPUUtilization(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func MemoryContainerEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.MemoryContainer(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func MemoryEdgeEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.MemoryEdge(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func MemoryContainerPeakEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.MemoryContainerPeak(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func MemoryEdgePeakEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.MemoryEdgePeak(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func MemoryContainerLimitSecondsEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.MemoryContainerReachLimitSeconds(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOEdgeTimeEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOEdgeTime(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOContainerTimeEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOContainerTime(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOEdgeReadEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOReadEdgeBytes(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOContainerReadEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOReadContainerBytes(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOEdgeWriteEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOWriteEdgeBytes(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOContainerWriteEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOWriteContainerBytes(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOFilesystemUsedEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOFilesystemUsedBytes(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func IOFilesystemSizeEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.IOFilesystemSizeBytes(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeRxBytesEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgeRxBytes(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeRxPacketsEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgeRxPackets(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeRxDroppedEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgePacketRxDropped(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeRxErrorEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgeRxError(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeTxBytesEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgeTxBytes(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeTxPacketsEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgeTxPackets(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeTxDroppedEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgePacketTxDropped(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetEdgeTxErrorEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkEdgeTxError(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerRxBytesEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerRxBytes(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerRxPacketsEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerRxPackets(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerRxDroppedEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerPacketRxDropped(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerRxErrorEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerRxError(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerTxBytesEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerTxBytes(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerTxPacketsEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerTxPackets(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerTxDroppedEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerPacketTxDropped(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func NetContainerTxErrorEP(requestId string, args map[string]interface{}) []byte {
	containerId, timestamp, err := parseMonitorContainerArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.NetworkContainerTxError(*containerId, *timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

func ThermalsEP(requestId string, args map[string]interface{}) []byte {
	timestamp, err := parseMonitorEdgeArgs(args)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	result, err := monitor2.Thermals(*timestamp)
	if err != nil {
		return replyMonitorError(requestId, err)
	}
	return replyMonitorResponse(requestId, *result)
}

//Process monitoring commands only
func parseMonitor(jsonMsg map[string]interface{}) {
	requestId, ok := jsonMsg["requestId"].(string)
	if !ok {
		//We can't figure out what the request is. Ignore completely.
		return
	}
	args, ok := jsonMsg["args"].(map[string]interface{})
	if !ok {
		replyReject(requestId)
		return
	}
	replyAck(requestId, "monitor")
	//As deployments can take a long time, these operations are done in a separate go function so that the agent can process other requests in parallel
	//Also the RabbitMQ server will disconnect the agent if the client does not process the message on time
	go func() {
		var response []byte
		switch jsonMsg["command"] {
		//CPU average usage per core over the last 10 seconds
		case "cpu_edge_avg":
			response = CPUEdgeAvgEP(requestId, args)
		case "cpu_container_avg":
			response = CPUContainerAvgEP(requestId, args)
		//CPU time by core
		case "cpu_time":
			response = CPUTimeEP(requestId, args)
		//CPU overall utilization over time
		case "cpu_utilization":
			response = CPUUtilizeEP(requestId, args)
		//Memory usage in bytes
		case "memory_container":
			response = MemoryContainerEP(requestId, args)
		case "memory_edge":
			response = MemoryEdgeEP(requestId, args)
		//Maximum memory usage in bytes
		case "memory_container_peak":
			response = MemoryContainerPeakEP(requestId, args)
		case "memory_edge_peak":
			response = MemoryEdgePeakEP(requestId, args)
		//Seconds of the container reaching its memory soft limits
		case "memory_container_limit_seconds":
			response = MemoryContainerLimitSecondsEP(requestId, args)
		//Time spent in io
		case "io_edge_time":
			response = IOEdgeTimeEP(requestId, args)
		case "io_container_time":
			response = IOContainerTimeEP(requestId, args)
		//Bytes read in io
		case "io_edge_read":
			response = IOEdgeReadEP(requestId, args)
		case "io_container_read":
			response = IOContainerReadEP(requestId, args)
		//Bytes written in io
		case "io_edge_write":
			response = IOEdgeWriteEP(requestId, args)
		case "io_container_write":
			response = IOContainerWriteEP(requestId, args)
		//Filesystem usage in bytes
		case "io_filesystem_used":
			response = IOFilesystemUsedEP(requestId, args)
		//Total filesystem size in bytes
		case "io_filesystem_size":
			response = IOFilesystemSizeEP(requestId, args)
		//Edge side network queries
		//Received bytes
		case "net_edge_rx_bytes":
			response = NetEdgeRxBytesEP(requestId, args)
		//Received packets
		case "net_edge_rx_packets":
			response = NetEdgeRxPacketsEP(requestId, args)
		//Received packets dropped
		case "net_edge_rx_dropped":
			response = NetEdgeRxDroppedEP(requestId, args)
		//Received errors
		case "net_edge_rx_error":
			response = NetEdgeRxErrorEP(requestId, args)
		//Transmitted bytes
		case "net_edge_tx_bytes":
			response = NetEdgeTxBytesEP(requestId, args)
		//Transmitted packets
		case "net_edge_tx_packets":
			response = NetEdgeTxPacketsEP(requestId, args)
		//Transmitted packets dropped
		case "net_edge_tx_dropped":
			response = NetEdgeTxDroppedEP(requestId, args)
		//Transmission errors
		case "net_edge_tx_error":
			response = NetEdgeTxErrorEP(requestId, args)
		//Container side network queries. Same functionality as the edge side.
		case "net_container_rx_bytes":
			response = NetContainerRxBytesEP(requestId, args)
		case "net_container_rx_packets":
			response = NetContainerRxPacketsEP(requestId, args)
		case "net_container_rx_dropped":
			response = NetContainerRxDroppedEP(requestId, args)
		case "net_container_rx_error":
			response = NetContainerRxErrorEP(requestId, args)
		case "net_container_tx_bytes":
			response = NetContainerTxBytesEP(requestId, args)
		case "net_container_tx_packets":
			response = NetContainerTxPacketsEP(requestId, args)
		case "net_container_tx_dropped":
			response = NetContainerTxDroppedEP(requestId, args)
		case "net_container_tx_error":
			response = NetContainerTxErrorEP(requestId, args)
		case "thermal":
			response = ThermalsEP(requestId, args)
		default:
			response = replyMonitorError(requestId, errors.New("unknown command"))
		}
		err := ch.Publish(
			"",
			responseQueue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        response,
			},
		)
		if err != nil {
			log.Error.Println("Failed pushing metric response")
			log.Error.Println(err)
		}
	}()
}

//Parsing arguments
func parseMonitorContainerArgs(args map[string]interface{}) (*string, *time.Time, error) {
	containerId, ok := args["containerId"].(string)
	if !ok {
		return nil, nil, errors.New("monitor - cannot parse request arguments")
	}
	timestampRaw, ok := args["time"].(int64)
	if !ok {
		return nil, nil, errors.New("monitor - cannot parse request arguments")
	}
	timestamp := time.Unix(timestampRaw, 0)
	return &containerId, &timestamp, nil
}

func parseMonitorEdgeArgs(args map[string]interface{}) (*time.Time, error) {
	timestampRaw, ok := args["time"].(int64)
	if !ok {
		return nil, errors.New("monitor - cannot parse request arguments")
	}
	timestamp := time.Unix(timestampRaw, 0)
	return &timestamp, nil
}

func replyMonitorError(requestId string, err error) []byte {
	return replyError(requestId, "monitor", err)
}

func replyMonitorResponse(requestId string, metric monitor2.Metric) []byte {
	response, _ := json.Marshal(map[string]interface{}{
		"requestId": requestId,
		"status":    "ok",
		"api":       "monitor",
		"metric":    metric,
	})
	return response
}

func replyReject(requestId string) {
	response := replyMonitorError(requestId, errors.New("monitoring api not enabled as prometheus is not deployed"))
	_ = ch.Publish(
		"",
		responseQueue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        response,
		},
	)
}
