package monitor

import (
	"fmt"
	"time"
)

//Network related queries
//You will need to specify the agent ID (which refers to an edge device), container ID (If applicable), and time.

//Bytes received on the edge device over the past 10 seconds
func NetworkEdgeRxBytes(time time.Time) (*Metric, error) {
	const query = "rate(node_network_receive_bytes_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Bytes received on the container over the past 10 seconds
func NetworkContainerRxBytes(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_receive_bytes_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Bytes transmitted on the edge device over the past 10 seconds
func NetworkEdgeTxBytes(time time.Time) (*Metric, error) {
	const query = "rate(node_network_transmit_bytes_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Bytes transmitted on the container over the past 10 seconds
func NetworkContainerTxBytes(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_transmit_bytes_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Packets received on the edge device over the past 10 seconds
func NetworkEdgeRxPackets(time time.Time) (*Metric, error) {
	const query = "rate(node_network_receive_packets_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Packets received on the container over the past 10 seconds
func NetworkContainerRxPackets(containerId string, time time.Time) (*Metric, error) {
	const query = "container_network_receive_packets_total{id='/docker/%s'}"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Packets transmitted on the edge device over the past 10 seconds
func NetworkEdgeTxPackets(time time.Time) (*Metric, error) {
	const query = "rate(node_network_transmit_packets_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Packets transmitted on the container over the past 10 seconds
func NetworkContainerTxPackets(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_transmit_packets_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Packets dropped during receiving on the edge device over the past 10 seconds
func NetworkEdgePacketRxDropped(time time.Time) (*Metric, error) {
	const query = "rate(node_network_receive_drop_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Packets dropped during receiving on the container over the past 10 seconds
func NetworkContainerPacketRxDropped(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_receive_packets_dropped_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Packets dropped during transmitting on the edge device over the past 10 seconds
func NetworkEdgePacketTxDropped(time time.Time) (*Metric, error) {
	const query = "rate(node_network_transmit_drop_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Packets dropped during transmitting on the container over the past 10 seconds
func NetworkContainerPacketTxDropped(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_transmit_packets_dropped_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Errors occurred during receiving on the edge device over the past 10 seconds
func NetworkEdgeRxError(time time.Time) (*Metric, error) {
	const query = "rate(node_network_receive_errs_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Errors occurred during receiving on the container over the past 10 seconds
func NetworkContainerRxError(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_receive_errors_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

//Errors occurred during transmitting on the edge device over the past 10 seconds
func NetworkEdgeTxError(time time.Time) (*Metric, error) {
	const query = "rate(node_network_transmit_errs_total{device!='lo'}[10s])"
	metric, err := restQuery(query, time)
	if err != nil {
		return nil, err
	}
	return metric, err
}

//Errors occurred during transmitting on the container over the past 10 seconds
func NetworkContainerTxError(containerId string, time time.Time) (*Metric, error) {
	const query = "rate(container_network_transmit_errors_total{id='/docker/%s'}[10s])"
	metric, err := restQuery(fmt.Sprintf(query, containerId), time)
	if err != nil {
		return nil, err
	}
	return metric, nil
}
