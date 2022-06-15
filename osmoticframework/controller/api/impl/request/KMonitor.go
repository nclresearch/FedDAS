package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"osmoticframework/controller/api/impl/request/monitor/query"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types/metric"
	"osmoticframework/controller/vars"
	"strconv"
	"time"
)

/*
All Kubernetes monitoring API functions
Prometheus smallest scale in time is second. Therefore there is no point in using nanosecond precision.
Note: Queries to node level metric must use the node's host IP. You can obtain one by using KEndpointInfo function
*/

//Node (Individual server host) average usage per core from the last 10 seconds (Separated by core ID)
func KCPUCoreAvg(nodeName string, time time.Time) ([]metric.KCPUUsage, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KCPUNodeAvg, nodeName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KCPUUsage, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var cpuUsage = metric.KCPUUsage{}
		core, _ := strconv.Atoi(m.Key["cpu"])
		cpuUsage.Core = core
		cpuUsage.Node = m.Key["node"]
		cpuUsage.Usage = m.Scalar.Value
		ret = append(ret, cpuUsage)
	}
	return ret, nil
}

//Pod average usage (includes all replicas) from the last 10 seconds
func KCPUPodAvg(podName string, time time.Time) ([]metric.KCPUPodUsage, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KCPUPodAvg, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KCPUPodUsage, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var cpuUsage = metric.KCPUPodUsage{}
		cpuUsage.Pod = m.Key["pod"]
		cpuUsage.Usage = m.Scalar.Value
		ret = append(ret, cpuUsage)
	}
	return ret, nil
}

//Cumulative CPU time by core
func KCPUTime(nodeName string, time time.Time) ([]metric.KCPUTime, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KCPUTime, nodeName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KCPUTime, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var cpuTime = metric.KCPUTime{}
		core, _ := strconv.Atoi(m.Key["cpu"])
		cpuTime.Core = core
		cpuTime.Node = m.Key["node"]
		cpuTime.CPUTime = m.Scalar.Value
		ret = append(ret, cpuTime)
	}
	return ret, nil
}

//Overall node CPU utilization
func KCPUUtilization(nodeName string, time time.Time) (*metric.KCPUUsage, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KCPUUtil, nodeName), time)
	if err != nil {
		return nil, err
	}
	var ret = metric.KCPUUsage{}
	for _, m := range promMetric.Data.([]metric.Vector) {
		ret.Node = m.Key["node"]
		ret.Usage = m.Scalar.Value
	}
	return &ret, nil
}

//Memory usage in bytes in a pod
func KMemoryPod(podName string, time time.Time) (*metric.KMemoryPodMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KMemoryPod, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = metric.KMemoryPodMetric{}
	for _, m := range promMetric.Data.([]metric.Vector) {
		ret.Pod = m.Key["pod"]
		ret.UsageBytes = int64(m.Scalar.Value)
	}
	return &ret, nil
}

//Memory usage in bytes in the whole cluster
func KMemoryNode(nodeIP string, time time.Time) (*metric.KMemoryNodeMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KMemoryNode, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = metric.KMemoryNodeMetric{}
	for _, m := range promMetric.Data.([]metric.Vector) {
		ret.Node = m.Key["node"]
		ret.UsageBytes = int64(m.Scalar.Value)
	}
	return &ret, nil
}

//Highest memory usage in bytes over the past 5 minutes in a pod
func KMemoryPodPeak(podName string, time time.Time) (*metric.KMemoryPodPeakMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KMemoryPodPeak, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = metric.KMemoryPodPeakMetric{}
	for _, m := range promMetric.Data.([]metric.Vector) {
		ret.Pod = m.Key["pod"]
		ret.UsageBytes = int64(m.Scalar.Value)
	}
	return &ret, nil
}

//Highest memory usage in bytes over the past 5 minutes in a node
func KMemoryNodePeak(nodeIP string, time time.Time) (*metric.KMemoryNodePeakMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KMemoryNodePeak, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = metric.KMemoryNodePeakMetric{}
	for _, m := range promMetric.Data.([]metric.Vector) {
		ret.Node = m.Key["node"]
		ret.UsageBytes = int64(m.Scalar.Value)
	}
	return &ret, nil
}

//The time in seconds a pod reaches the memory soft limit over the past 5 minutes (It cannot exceed 300 seconds)
func KMemoryPodReachLimitSeconds(podName string, time time.Time) (*metric.KMemoryPodReachLimitSecondsMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KSoftLimitQuery, podName), time)
	if err != nil {
		return nil, err
	}
	//Check if promMetric is empty
	if len(promMetric.Data.([]metric.Vector)) == 0 {
		return nil, nil
	}
	//Check if soft limit is enabled in the pod. If not set, return error
	if promMetric.Data.([]metric.Vector)[0].Scalar.Value == 0 {
		return nil, errors.New("soft limit for container is not set")
	}
	promMetric, err = restQuery(fmt.Sprintf(query.KMemoryPodReachLimitSeconds, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = metric.KMemoryPodReachLimitSecondsMetric{}
	for _, m := range promMetric.Data.([]metric.Vector) {
		ret.Pod = m.Key["pod"]
		ret.MemoryPressureSeconds = int64(m.Scalar.Value)
	}
	return &ret, nil
}

//Cumulative number of seconds spent doing IO operations on a node
func KIONodeTime(nodeIP string, time time.Time) ([]metric.KIONodeTimeMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIONodeTime, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIONodeTimeMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioTime = metric.KIONodeTimeMetric{}
		ioTime.Device = m.Key["device"]
		ioTime.Node = m.Key["node"]
		ioTime.Time = m.Scalar.Value
		ret = append(ret, ioTime)
	}
	return ret, nil
}

//Cumulative number of seconds spent doing IO operations on a pod
func KIOPodTime(podName string, time time.Time) ([]metric.KIOPodTimeMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOPodTime, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIOPodTimeMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioTime = metric.KIOPodTimeMetric{}
		ioTime.Device = m.Key["device"]
		ioTime.Pod = m.Key["pod"]
		ioTime.Time = m.Scalar.Value
		ret = append(ret, ioTime)
	}
	return ret, nil
}

//Cumulative amount of read bytes on the node
func KIOReadNodeBytes(nodeIP string, time time.Time) ([]metric.KIONodeReadMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOReadNodeBytes, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIONodeReadMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioRead = metric.KIONodeReadMetric{}
		ioRead.Device = m.Key["device"]
		ioRead.Node = m.Key["node"]
		ioRead.BytesRead = int64(m.Scalar.Value)
		ret = append(ret, ioRead)
	}
	return ret, nil
}

//Cumulative amount of read bytes on a pod
func KIOReadPodBytes(podName string, time time.Time) ([]metric.KIOPodReadMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOReadPodBytes, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIOPodReadMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioRead = metric.KIOPodReadMetric{}
		ioRead.Device = m.Key["device"]
		ioRead.Pod = m.Key["pod"]
		ioRead.BytesRead = int64(m.Scalar.Value)
		ret = append(ret, ioRead)
	}
	return ret, nil
}

//Cumulative amount of write bytes on the node
func KIOWriteNodeBytes(nodeIP string, time time.Time) ([]metric.KIONodeWriteMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOWriteNodeBytes, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIONodeWriteMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioWrite = metric.KIONodeWriteMetric{}
		ioWrite.Device = m.Key["device"]
		ioWrite.Node = m.Key["node"]
		ioWrite.BytesWritten = int64(m.Scalar.Value)
		ret = append(ret, ioWrite)
	}
	return ret, nil
}

//Cumulative amount of write bytes on a pod
func KIOWritePodBytes(podName string, time time.Time) ([]metric.KIOPodWriteMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOWritePodBytes, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIOPodWriteMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioWrite = metric.KIOPodWriteMetric{}
		ioWrite.Device = m.Key["device"]
		ioWrite.Pod = m.Key["pod"]
		ioWrite.BytesWritten = int64(m.Scalar.Value)
		ret = append(ret, ioWrite)
	}
	return ret, nil
}

//Total disk space used in bytes
func KIOFilesystemUsedBytes(nodeIP string, time time.Time) ([]metric.KIOFilesystemUsedMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOFilesystemUsedBytes, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIOFilesystemUsedMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioUsed = metric.KIOFilesystemUsedMetric{}
		ioUsed.Device = m.Key["device"]
		ioUsed.Node = m.Key["node"]
		ioUsed.BytesUsed = int64(m.Scalar.Value)
		ioUsed.Mountpoint = m.Key["mountpoint"]
		ret = append(ret, ioUsed)
	}
	return ret, nil
}

//Filesystem size in bytes
func KIOFilesystemSizeBytes(nodeIP string, time time.Time) ([]metric.KIOFilesystemSizeMetric, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KIOFilesystemSizeBytes, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KIOFilesystemSizeMetric, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var ioSize = metric.KIOFilesystemSizeMetric{}
		ioSize.Device = m.Key["device"]
		ioSize.Node = m.Key["node"]
		ioSize.BytesSize = int64(m.Scalar.Value)
		ioSize.Mountpoint = m.Key["mountpoint"]
		ret = append(ret, ioSize)
	}
	return ret, nil
}

//Network receive bytes on the node
func KNetworkNodeRxBytes(nodeIP string, time time.Time) ([]metric.KNodeNetworkBytes, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeRxBytes, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkBytes, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KNodeNetworkBytes{}
		netRx.Node = m.Key["node"]
		netRx.Interface = m.Key["device"]
		netRx.Bytes = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network receive bytes on a pod
func KNetworkPodRxBytes(podName string, time time.Time) ([]metric.KPodNetworkBytes, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodRxBytes, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkBytes, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KPodNetworkBytes{}
		netRx.Pod = m.Key["pod"]
		netRx.Interface = m.Key["device"]
		netRx.Bytes = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network transmit bytes on the node
func KNetworkNodeTxBytes(nodeIP string, time time.Time) ([]metric.KNodeNetworkBytes, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeTxBytes, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkBytes, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KNodeNetworkBytes{}
		netTx.Node = m.Key["node"]
		netTx.Interface = m.Key["device"]
		netTx.Bytes = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network transmit bytes on a pod
func KNetworkPodTxBytes(podName string, time time.Time) ([]metric.KPodNetworkBytes, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodTxBytes, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkBytes, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KPodNetworkBytes{}
		netTx.Pod = m.Key["pod"]
		netTx.Interface = m.Key["device"]
		netTx.Bytes = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network received packets on the node
func KNetworkNodeRxPackets(nodeIP string, time time.Time) ([]metric.KNodeNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeRxPackets, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KNodeNetworkPackets{}
		netRx.Node = m.Key["node"]
		netRx.Interface = m.Key["device"]
		netRx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network received packets on a pod
func KNetworkPodRxPackets(podName string, time time.Time) ([]metric.KPodNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodRxPackets, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KPodNetworkPackets{}
		netRx.Pod = m.Key["pod"]
		netRx.Interface = m.Key["device"]
		netRx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network transmit packets on the node
func KNetworkNodeTxPackets(nodeIP string, time time.Time) ([]metric.KNodeNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeTxPackets, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KNodeNetworkPackets{}
		netTx.Node = m.Key["node"]
		netTx.Interface = m.Key["device"]
		netTx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network transmit packets on a pod
func KNetworkPodTxPackets(podName string, time time.Time) ([]metric.KPodNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodTxPackets, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KPodNetworkPackets{}
		netTx.Pod = m.Key["pod"]
		netTx.Interface = m.Key["device"]
		netTx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network dropped received packets on the node
func KNetworkNodeRxDropped(nodeIP string, time time.Time) ([]metric.KNodeNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeRxPacketDropped, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KNodeNetworkPackets{}
		netRx.Node = m.Key["node"]
		netRx.Interface = m.Key["device"]
		netRx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network dropped received packets on a pod
func KNetworkPodRxDropped(podName string, time time.Time) ([]metric.KPodNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodRxPacketDropped, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KPodNetworkPackets{}
		netRx.Pod = m.Key["pod"]
		netRx.Interface = m.Key["device"]
		netRx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network dropped transmit packets on the node
func KNetworkNodeTxDropped(nodeIP string, time time.Time) ([]metric.KNodeNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeTxPacketDropped, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KNodeNetworkPackets{}
		netTx.Node = m.Key["node"]
		netTx.Interface = m.Key["device"]
		netTx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network dropped transmit packets on a pod
func KNetworkPodTxDropped(podName string, time time.Time) ([]metric.KPodNetworkPackets, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodTxPacketDropped, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkPackets, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KPodNetworkPackets{}
		netTx.Pod = m.Key["pod"]
		netTx.Interface = m.Key["device"]
		netTx.Packets = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network receive errors on the node
func KNetworkNodeRxError(nodeIP string, time time.Time) ([]metric.KNodeNetworkErrors, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeRxError, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkErrors, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KNodeNetworkErrors{}
		netRx.Node = m.Key["node"]
		netRx.Interface = m.Key["device"]
		netRx.Errors = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network receive errors on a pod
func KNetworkPodRxError(podName string, time time.Time) ([]metric.KPodNetworkErrors, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodRxError, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KPodNetworkErrors, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netRx = metric.KPodNetworkErrors{}
		netRx.Pod = m.Key["pod"]
		netRx.Interface = m.Key["device"]
		netRx.Errors = uint64(m.Scalar.Value)
		ret = append(ret, netRx)
	}
	return ret, nil
}

//Network transmit errors on the node
func KNetworkNodeTxError(nodeIP string, time time.Time) ([]metric.KNodeNetworkErrors, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkNodeTxError, nodeIP), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkErrors, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KNodeNetworkErrors{}
		netTx.Node = m.Key["node"]
		netTx.Interface = m.Key["device"]
		netTx.Errors = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Network transmit errors on a pod
func KNetworkPodTxError(podName string, time time.Time) ([]metric.KNodeNetworkErrors, error) {
	promMetric, err := restQuery(fmt.Sprintf(query.KNetworkPodTxError, podName), time)
	if err != nil {
		return nil, err
	}
	var ret = make([]metric.KNodeNetworkErrors, 0)
	for _, m := range promMetric.Data.([]metric.Vector) {
		var netTx = metric.KNodeNetworkErrors{}
		netTx.Node = m.Key["node"]
		netTx.Interface = m.Key["device"]
		netTx.Errors = uint64(m.Scalar.Value)
		ret = append(ret, netTx)
	}
	return ret, nil
}

//Endpoint IP and node host IP
//This returns the relationship between the node exporter's endpoint IP and the node's hostname
//You can use this to find metric on the host level
func KEndpointInfo() (*metric.PromMetric, error) {
	promMetric, err := restQuery(query.KEndpointInfo, time.Now())
	if err != nil {
		return nil, err
	}
	return promMetric, nil
}

//Queries in a specific point in time.
func restQuery(query string, time time.Time) (*metric.PromMetric, error) {
	if vars.GetPrometheusAddress() == "" {
		log.Fatal.Panicln("Prometheus address not defined!")
	}
	client := resty.New()
	response, err := client.R().
		SetQueryParams(map[string]string{
			"query": query,
			"time":  strconv.FormatInt(time.Unix(), 10),
		}).
		Get(vars.GetPrometheusAddress() + "/query")
	if err != nil {
		return nil, err
	}
	err = processStatus(response)
	if err != nil {
		return nil, err
	}
	promMetric, err := parsePromResponse(response.Body())
	if err != nil {
		return nil, err
	}
	return promMetric, nil
}

//Queries in a specific range in time, and it will return multiple points in time
//Prometheus smallest scale in time is second. Therefore there is no point in using nanosecond precision.
func restQueryRange(query string, from, to time.Time, step time.Duration) (*metric.PromMetric, error) {
	if vars.GetPrometheusAddress() == "" {
		log.Fatal.Panicln("Prometheus address not defined!")
	}
	client := resty.New()
	response, err := client.R().
		SetQueryParams(map[string]string{
			"query": query,
			"from":  strconv.FormatInt(from.Unix(), 10),
			"to":    strconv.FormatInt(to.Unix(), 10),
			"step":  strconv.Itoa(int(step.Seconds())),
		}).
		Get(vars.GetPrometheusAddress() + "/query_range")
	if err != nil {
		return nil, err
	}
	err = processStatus(response)
	if err != nil {
		return nil, err
	}
	promMetric, err := parsePromResponse(response.Body())
	if err != nil {
		return nil, err
	}
	return promMetric, nil
}

func processStatus(response *resty.Response) error {
	if response.StatusCode() == 200 {
		return nil
	}
	var errorJson map[string]interface{}
	err := json.Unmarshal(response.Body(), &errorJson)
	if err != nil {
		return err
	}
	//API errors
	errString, ok := errorJson["error"].(string)
	if !ok {
		return errors.New("unknown error")
	}
	return errors.New(errString)
}

func parsePromResponse(response []byte) (*metric.PromMetric, error) {
	var responseJson map[string]interface{}
	err := json.Unmarshal(response, &responseJson)
	if err != nil {
		return nil, errors.New("unable to parse broken json from prometheus")
	}
	var promMetric metric.PromMetric
	resultType, ok := responseJson["data"].(map[string]interface{})["resultType"].(string)
	if !ok {
		return nil, errors.New("cannot parse api response")
	}
	resultRaw := responseJson["data"].(map[string]interface{})["result"]
	promMetric = metric.PromMetric{}
	switch resultType {
	case "matrix":
		var result []map[string]interface{}
		err := mapstructure.Decode(resultRaw, &result)
		if err != nil {
			return nil, err
		}
		promMetric.Type = metric.MatrixType
		var matrix = make([]metric.Matrix, 0)
		for _, dev := range result {
			devInfoDecoded, ok := dev["promMetric"].(map[string]interface{})
			//mapstructure has decoded some of the fields in promMetric to int64 or float64 instead of string.
			//To minimize the unnecessary type assertions, we'll convert all of them to string
			devInfo := make(map[string]string)
			for key, value := range devInfoDecoded {
				switch value.(type) {
				case string:
					devInfo[key] = value.(string)
				case int64:
					devInfo[key] = strconv.FormatInt(value.(int64), 10)
				case float64:
					devInfo[key] = strconv.FormatFloat(value.(float64), 'E', -1, 64)
				}
			}
			if !ok {
				return nil, errors.New("cannot parse api response")
			}
			//Assertion must be interface due to mixed types (int and string)
			//JSON structure of matrix
			/*
				"result": [
					{
						"promMetric": {
							"cpu": "0"
						},
						"values": [
							[
								1574172422,
								"0.028000000000001805"
							],
							[
								1574172423,
								"0.028000000000001805"
							]
						]
					}
				]
			*/
			var scalars [][]interface{}
			err := mapstructure.Decode(dev["values"], &scalars)
			if err != nil {
				return nil, err
			}
			entry := metric.Matrix{Key: devInfo}
			values := make([]metric.Scalar, 0)
			for _, scalar := range scalars {
				value, err := parseScalar(scalar)
				if err != nil {
					return nil, err
				}
				values = append(values, *value)
			}
			entry.Values = values
			matrix = append(matrix, entry)
		}
		promMetric.Data = matrix
	case "vector":
		var result []map[string]interface{}
		err := mapstructure.Decode(resultRaw, &result)
		if err != nil {
			return nil, err
		}
		promMetric.Type = metric.VectorType
		var vectors = make([]metric.Vector, 0)
		for _, dev := range result {
			devInfoDecoded, ok := dev["promMetric"].(map[string]interface{})
			//mapstructure has decoded some of the fields in promMetric to int64 or float64 instead of string.
			//To minimize the unnecessary type assertions, we'll convert all of them to string
			devInfo := make(map[string]string)
			for key, value := range devInfoDecoded {
				switch value.(type) {
				case string:
					devInfo[key] = value.(string)
				case int64:
					devInfo[key] = strconv.FormatInt(value.(int64), 10)
				case float64:
					devInfo[key] = strconv.FormatFloat(value.(float64), 'E', -1, 64)
				}
			}
			if !ok {
				return nil, errors.New("cannot parse api response")
			}
			//Assertion must be interface due to mixed types (int and string)
			//JSON structure of vectors
			/*
				"result": [
					{
						"promMetric": {
							"cpu": "4"
						},
						"value": [
							1574172422,
							"0.027999999999999935"
						]
					}
				]
			*/
			//There's only one element in "value" for vectors. So no 2-dimensional arrays
			scalar, ok := dev["value"].([]interface{})
			if !ok {
				return nil, errors.New("cannot parse api response")
			}
			value, err := parseScalar(scalar)
			if err != nil {
				return nil, err
			}
			entry := metric.Vector{
				Key:    devInfo,
				Scalar: *value,
			}
			vectors = append(vectors, entry)
		}
		promMetric.Data = vectors
	case "scalar":
		//Assertion must be interface due to mixed types (int and string)
		//JSON structure of scalars
		/*
			"result": [
				1574172422,
				"0.027999999999999935"
			]
		*/
		var result []interface{}
		err := mapstructure.Decode(resultRaw, &result)
		if err != nil {
			return nil, err
		}
		promMetric.Type = metric.ScalarType
		value, err := parseScalar(result)
		if err != nil {
			return nil, err
		}
		promMetric.Data = *value
	case "string":
		//Unused in Prometheus. So do nothing.
		promMetric.Type = metric.StringType
	default:
		return nil, errors.New("unknown data type")
	}
	return &promMetric, nil
}

//Parses a metric value to either a float or undefined
func parseScalar(scalar []interface{}) (*metric.Scalar, error) {
	var timestamp time.Time
	//Dynamic typing shenanigans from json
	//Sometimes Prometheus returns milliseconds in its timestamp, which makes the timestamp a float
	//Sometimes it does not, which makes it an int
	switch scalar[0].(type) {
	case int64:
		unixTimestamp, _ := scalar[0].(int64)
		timestamp = time.Unix(unixTimestamp, 0)
	case float64:
		unixTimestamp, _ := scalar[0].(float64)
		sec := int64(unixTimestamp)
		//nano seconds
		nsec := int64(unixTimestamp - (float64(sec))*1000000)
		timestamp = time.Unix(sec, nsec)
	default:
		return nil, errors.New("cannot parse time")
	}
	rawValue := scalar[1].(string)
	if rawValue == "NaN" {
		return &metric.Scalar{
			Time:      timestamp,
			Value:     0,
			Undefined: true,
		}, nil
	} else {
		floatValue, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil, errors.New("unable to parse scalar: " + rawValue)
		}
		return &metric.Scalar{
			Time:      timestamp,
			Value:     floatValue,
			Undefined: false,
		}, nil
	}
}
