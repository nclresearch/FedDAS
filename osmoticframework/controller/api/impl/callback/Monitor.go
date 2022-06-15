package callback

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types/metric"
	"strconv"
)

//Reads the response message from the agent.
//See README.md for the message structure
func ParseMonitor(message map[string]interface{}) {
	requestId, ok := message["requestId"].(string)
	if !ok {
		//Cannot figure out request content. Ignore
		return
	}
	_requestTask, exists := request.MonitorRequests.Load(requestId)
	if !exists {
		//Request ID does not exist in memory. Ignore
		return
	}
	requestTask := _requestTask.(request.ImplRequestTask)
	status, ok := message["status"].(string)
	if !ok {
		//No status. Ignore
		return
	}
	switch status {
	case "ack":
		req := request.ImplRequestTask{
			AgentId: requestTask.AgentId,
			API:     requestTask.API,
			Command: requestTask.Command,
			Ack:     true,
			Time:    requestTask.Time,
			Args:    requestTask.Args,
			Timeout: requestTask.Timeout,
		}
		request.MonitorRequests.Store(requestId, req)
	case "ok":
		//Metrics requests
		//Deserialize metric
		if message["metric"] == nil {
			err := errors.New("invalid response")
			log.Error.Println("Agent sent no metric. Monitor request fail")
			log.Error.Println(err)
			CallbackError(requestId, err)
			return
		}
		var promMetric metric.PromMetric
		err := mapstructure.Decode(message["metric"], &promMetric)
		if err != nil {
			log.Error.Println("Failed decoding reply in monitoring API")
			log.Error.Println(err)
			CallbackError(requestId, err)
			return
		}
		CallbackOk(requestId, parseMetric(requestTask.AgentId, requestTask.Command, promMetric))
	case "failed":
		var err error
		errStr, ok := message["error"].(string)
		if !ok {
			err = errors.New("unknown error")
		} else {
			err = errors.New(errStr)
		}
		log.Error.Printf("%s (req: %s) >> Request failed\n", requestTask.AgentId, requestId)
		log.Error.Println(err)
		CallbackError(requestId, err)
		request.MonitorRequests.Delete(requestId)
	}
}

func parseMetric(agentId, command string, promMetric metric.PromMetric) interface{} {
	switch command {
	case "cpu_edge_avg":
		ret := make([]metric.CpuEdgeMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			cpuEdgeMetric := metric.CpuEdgeMetric{}
			cpuEdgeMetric.Core, _ = strconv.Atoi(m.Key["cpu"])
			cpuEdgeMetric.Agent = agentId
			cpuEdgeMetric.Usage = m.Scalar.Value
			ret = append(ret, cpuEdgeMetric)
		}
		return ret
	case "cpu_container_avg":
		ret := make([]metric.CpuContainerMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			cpuContainerMetric := metric.CpuContainerMetric{}
			cpuContainerMetric.Container = m.Key["container"][8:]
			cpuContainerMetric.Core, _ = strconv.Atoi(m.Key["cpu"][3:])
			cpuContainerMetric.Agent = agentId
			cpuContainerMetric.Usage = m.Scalar.Value
			ret = append(ret, cpuContainerMetric)
		}
		return ret
	case "cpu_time":
		ret := make([]metric.CpuEdgeTimeMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			cpuEdgeTimeMetric := metric.CpuEdgeTimeMetric{}
			cpuEdgeTimeMetric.Core, _ = strconv.Atoi(m.Key["cpu"][3:])
			cpuEdgeTimeMetric.Agent = agentId
			cpuEdgeTimeMetric.Time = m.Scalar.Value
			ret = append(ret, cpuEdgeTimeMetric)
		}
		return ret
	case "cpu_utilization":
		ret := metric.CpuEdgeOverallUsageMetric{}
		ret.Agent = agentId
		ret.Usage = promMetric.Data.(metric.Vector).Scalar.Value
		return ret
	case "memory_container":
		ret := metric.MemoryContainerMetric{}
		ret.Container = promMetric.Data.(metric.Vector).Key["container"][8:]
		ret.Agent = agentId
		ret.Usage = uint64(promMetric.Data.(metric.Vector).Scalar.Value)
		return ret
	case "memory_edge":
		ret := metric.MemoryEdgeMetric{}
		ret.Agent = agentId
		ret.Usage = uint64(promMetric.Data.(metric.Vector).Scalar.Value)
		return ret
	case "memory_container_peak":
		ret := metric.MemoryContainerMetric{}
		ret.Container = promMetric.Data.(metric.Vector).Key["container"][8:]
		ret.Agent = agentId
		ret.Usage = uint64(promMetric.Data.(metric.Vector).Scalar.Value)
		return ret
	case "memory_edge_peak":
		ret := metric.MemoryEdgeMetric{}
		ret.Agent = agentId
		ret.Usage = uint64(promMetric.Data.(metric.Vector).Scalar.Value)
		return ret
	case "memory_container_limit_seconds":
		ret := metric.MemoryContainerLimitSecondsMetric{}
		ret.Container = promMetric.Data.(metric.Vector).Key["container"][8:]
		ret.Agent = agentId
		ret.Time = uint64(promMetric.Data.(metric.Vector).Scalar.Value)
		return ret
	case "io_edge_time":
		ret := make([]metric.IOEdgeTimeMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioEdgeTimeMetric := metric.IOEdgeTimeMetric{}
			ioEdgeTimeMetric.Agent = agentId
			ioEdgeTimeMetric.Device = m.Key["device"]
			ioEdgeTimeMetric.Time = m.Scalar.Value
			ret = append(ret, ioEdgeTimeMetric)
		}
		return ret
	case "io_container_time":
		ret := make([]metric.IOContainerTimeMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioContainerTimeMetric := metric.IOContainerTimeMetric{}
			ioContainerTimeMetric.Container = m.Key["container"][8:]
			ioContainerTimeMetric.Device = m.Key["device"]
			ioContainerTimeMetric.Agent = agentId
			ioContainerTimeMetric.Time = m.Scalar.Value
			ret = append(ret, ioContainerTimeMetric)
		}
		return ret
	case "io_edge_read":
		ret := make([]metric.IOEdgeBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioEdgeBytesMetric := metric.IOEdgeBytesMetric{}
			ioEdgeBytesMetric.Agent = agentId
			ioEdgeBytesMetric.Device = m.Key["device"]
			ioEdgeBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, ioEdgeBytesMetric)
		}
		return ret
	case "io_container_read":
		ret := make([]metric.IOContainerBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioContainerBytesMetric := metric.IOContainerBytesMetric{}
			ioContainerBytesMetric.Container = m.Key["container"][8:]
			ioContainerBytesMetric.Device = m.Key["device"]
			ioContainerBytesMetric.Agent = agentId
			ioContainerBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, ioContainerBytesMetric)
		}
		return ret
	case "io_edge_write":
		ret := make([]metric.IOEdgeBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioEdgeBytesMetric := metric.IOEdgeBytesMetric{}
			ioEdgeBytesMetric.Agent = agentId
			ioEdgeBytesMetric.Device = m.Key["device"]
			ioEdgeBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, ioEdgeBytesMetric)
		}
		return ret
	case "io_container_write":
		ret := make([]metric.IOContainerBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioContainerBytesMetric := metric.IOContainerBytesMetric{}
			ioContainerBytesMetric.Container = m.Key["container"][8:]
			ioContainerBytesMetric.Device = m.Key["device"]
			ioContainerBytesMetric.Agent = agentId
			ioContainerBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, ioContainerBytesMetric)
		}
		return ret
	case "io_filesystem_used":
		ret := make([]metric.IOFilesystemBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioFilesystemBytesMetric := metric.IOFilesystemBytesMetric{}
			ioFilesystemBytesMetric.Device = m.Key["device"]
			ioFilesystemBytesMetric.MountPoint = m.Key["mountpoint"]
			ioFilesystemBytesMetric.Agent = agentId
			ioFilesystemBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, ioFilesystemBytesMetric)
		}
		return ret
	case "io_filesystem_size":
		ret := make([]metric.IOFilesystemBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			ioFilesystemBytesMetric := metric.IOFilesystemBytesMetric{}
			ioFilesystemBytesMetric.Device = m.Key["device"]
			ioFilesystemBytesMetric.MountPoint = m.Key["mountpoint"]
			ioFilesystemBytesMetric.Agent = agentId
			ioFilesystemBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, ioFilesystemBytesMetric)
		}
		return ret
	case "net_edge_rx_bytes":
		ret := make([]metric.NetworkEdgeBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgeBytesMetric := metric.NetworkEdgeBytesMetric{}
			networkEdgeBytesMetric.Agent = agentId
			networkEdgeBytesMetric.Device = m.Key["device"]
			networkEdgeBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgeBytesMetric)
		}
		return ret
	case "net_container_rx_bytes":
		ret := make([]metric.NetworkContainerBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerBytesMetric := metric.NetworkContainerBytesMetric{}
			networkContainerBytesMetric.Container = m.Key["container"][8:]
			networkContainerBytesMetric.Device = m.Key["device"]
			networkContainerBytesMetric.Agent = agentId
			networkContainerBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerBytesMetric)
		}
		return ret
	case "net_edge_tx_bytes":
		ret := make([]metric.NetworkEdgeBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgeBytesMetric := metric.NetworkEdgeBytesMetric{}
			networkEdgeBytesMetric.Agent = agentId
			networkEdgeBytesMetric.Device = m.Key["device"]
			networkEdgeBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgeBytesMetric)
		}
		return ret
	case "net_container_tx_bytes":
		ret := make([]metric.NetworkContainerBytesMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerBytesMetric := metric.NetworkContainerBytesMetric{}
			networkContainerBytesMetric.Container = m.Key["container"][8:]
			networkContainerBytesMetric.Device = m.Key["device"]
			networkContainerBytesMetric.Agent = agentId
			networkContainerBytesMetric.Bytes = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerBytesMetric)
		}
		return ret
	case "net_edge_rx_packets":
		ret := make([]metric.NetworkEdgePacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgePacketsMetric := metric.NetworkEdgePacketsMetric{}
			networkEdgePacketsMetric.Agent = agentId
			networkEdgePacketsMetric.Device = m.Key["device"]
			networkEdgePacketsMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgePacketsMetric)
		}
		return ret
	case "net_container_rx_packets":
		ret := make([]metric.NetworkContainerPacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerPacketsMetric := metric.NetworkContainerPacketsMetric{}
			networkContainerPacketsMetric.Container = m.Key["container"][8:]
			networkContainerPacketsMetric.Device = m.Key["device"]
			networkContainerPacketsMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerPacketsMetric)
		}
		return ret
	case "net_edge_tx_packets":
		ret := make([]metric.NetworkEdgePacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgePacketsMetric := metric.NetworkEdgePacketsMetric{}
			networkEdgePacketsMetric.Agent = agentId
			networkEdgePacketsMetric.Device = m.Key["device"]
			networkEdgePacketsMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgePacketsMetric)
		}
		return ret
	case "net_container_tx_packets":
		ret := make([]metric.NetworkContainerPacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerPacketsMetric := metric.NetworkContainerPacketsMetric{}
			networkContainerPacketsMetric.Container = m.Key["container"][8:]
			networkContainerPacketsMetric.Device = m.Key["device"]
			networkContainerPacketsMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerPacketsMetric)
		}
		return ret
	case "net_edge_rx_errors":
		ret := make([]metric.NetworkEdgeErrorsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgeErrorsMetric := metric.NetworkEdgeErrorsMetric{}
			networkEdgeErrorsMetric.Agent = agentId
			networkEdgeErrorsMetric.Device = m.Key["device"]
			networkEdgeErrorsMetric.Errors = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgeErrorsMetric)
		}
		return ret
	case "net_container_rx_errors":
		ret := make([]metric.NetworkContainerErrorsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerErrorsMetric := metric.NetworkContainerErrorsMetric{}
			networkContainerErrorsMetric.Container = m.Key["container"][8:]
			networkContainerErrorsMetric.Device = m.Key["device"]
			networkContainerErrorsMetric.Errors = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerErrorsMetric)
		}
		return ret
	case "net_edge_tx_errors":
		ret := make([]metric.NetworkEdgeErrorsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgeErrorsMetric := metric.NetworkEdgeErrorsMetric{}
			networkEdgeErrorsMetric.Agent = agentId
			networkEdgeErrorsMetric.Device = m.Key["device"]
			networkEdgeErrorsMetric.Errors = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgeErrorsMetric)
		}
		return ret
	case "net_container_tx_errors":
		ret := make([]metric.NetworkContainerErrorsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerErrorsMetric := metric.NetworkContainerErrorsMetric{}
			networkContainerErrorsMetric.Container = m.Key["container"][8:]
			networkContainerErrorsMetric.Device = m.Key["device"]
			networkContainerErrorsMetric.Errors = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerErrorsMetric)
		}
		return ret
	case "net_edge_rx_dropped":
		ret := make([]metric.NetworkEdgePacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgeDroppedMetric := metric.NetworkEdgePacketsMetric{}
			networkEdgeDroppedMetric.Agent = agentId
			networkEdgeDroppedMetric.Device = m.Key["device"]
			networkEdgeDroppedMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgeDroppedMetric)
		}
		return ret
	case "net_container_rx_dropped":
		ret := make([]metric.NetworkContainerPacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerDroppedMetric := metric.NetworkContainerPacketsMetric{}
			networkContainerDroppedMetric.Container = m.Key["container"][8:]
			networkContainerDroppedMetric.Device = m.Key["device"]
			networkContainerDroppedMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerDroppedMetric)
		}
		return ret
	case "net_edge_tx_dropped":
		ret := make([]metric.NetworkEdgePacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkEdgeDroppedMetric := metric.NetworkEdgePacketsMetric{}
			networkEdgeDroppedMetric.Agent = agentId
			networkEdgeDroppedMetric.Device = m.Key["device"]
			networkEdgeDroppedMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkEdgeDroppedMetric)
		}
		return ret
	case "net_container_tx_dropped":
		ret := make([]metric.NetworkContainerPacketsMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			networkContainerDroppedMetric := metric.NetworkContainerPacketsMetric{}
			networkContainerDroppedMetric.Container = m.Key["container"][8:]
			networkContainerDroppedMetric.Device = m.Key["device"]
			networkContainerDroppedMetric.Packets = uint64(m.Scalar.Value)
			ret = append(ret, networkContainerDroppedMetric)
		}
		return ret
	case "thermal":
		ret := make([]metric.ThermalMetric, 0)
		for _, m := range promMetric.Data.([]metric.Vector) {
			thermalMetric := metric.ThermalMetric{}
			thermalMetric.Agent = agentId
			thermalMetric.ZoneName = m.Key["type"]
			thermalMetric.ZoneUUID = m.Key["zone"]
			thermalMetric.Temperature = m.Scalar.Value
			ret = append(ret, thermalMetric)
		}
		return ret
	default:
		return promMetric
	}
}
