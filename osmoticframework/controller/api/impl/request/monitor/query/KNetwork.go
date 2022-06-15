package query

//Network metric queries for Kubernetes

// KNetworkNodeRxBytes: Returns the total number of bytes received by the node over the past 10 seconds.
const KNetworkNodeRxBytes = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_bytes_total{device!='lo'}[10s]))"

// KNetworkPodRxBytes: Returns the total number of bytes received by the pod over the past 1 minute.
const KNetworkPodRxBytes = "rate(container_network_receive_bytes_total{pod='%s'}[1m])"

// KNetworkNodeTxBytes: Returns the total number of bytes transmitted by the node over the past 10 seconds.
const KNetworkNodeTxBytes = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_bytes_total{device!='lo'}[10s]))"

// KNetworkPodTxBytes: Returns the total number of bytes transmitted by the pod over the past 1 minute.
const KNetworkPodTxBytes = "rate(container_network_transmit_bytes_total{pod='%s'}[1m])"

// KNetworkNodeRxPackets: Returns the total number of packets received by the node over the past 10 seconds.
const KNetworkNodeRxPackets = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_packets_total{device!='lo'}[10s]))"

// KNetworkPodRxPackets: Returns the total number of packets received by the pod over the past 1 minute.
const KNetworkPodRxPackets = "rate(container_network_receive_packets_total{pod='%s'}[1m])"

// KNetworkNodeTxPackets: Returns the total number of packets transmitted by the node over the past 10 seconds.
const KNetworkNodeTxPackets = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_packets_total{device!='lo'}[10s]))"

// KNetworkPodTxPackets: Returns the total number of packets transmitted by the pod over the past 1 minute.
const KNetworkPodTxPackets = "rate(container_network_transmit_packets_total{pod='%s'}[1m])"

// KNetworkNodeRxPacketsDropped: Returns the total number of packets dropped by the node over the past 10 seconds.
const KNetworkNodeRxPacketDropped = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_drop_total{device!='lo'}[10s]))"

// KNetworkPodRxPacketsDropped: Returns the total number of packets dropped by the pod over the past 1 minute.
const KNetworkPodRxPacketDropped = "rate(container_network_receive_packets_dropped_total{pod='%s'}[1m])"

// KNetworkNodeTxPacketsDropped: Returns the total number of packets dropped by the node over the past 10 seconds.
const KNetworkNodeTxPacketDropped = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_drop_total{device!='lo'}[10s]))"

// KNetworkPodTxPacketsDropped: Returns the total number of packets dropped by the pod over the past 1 minute.
const KNetworkPodTxPacketDropped = "rate(container_network_transmit_packets_dropped_total{pod='%s'}[1m])"

// KNetworkNodeRxError: Returns the total number of errors received by the node over the past 10 seconds.
const KNetworkNodeRxError = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_errs_total{device!='lo'}[10s]))"

// KNetworkPodRxError: Returns the total number of errors received by the pod over the past 1 minute.
const KNetworkPodRxError = "rate(container_network_receive_errors_total{pod='%s'}[1m])"

// KNetworkNodeTxError: Returns the total number of errors transmitted by the node over the past 10 seconds.
const KNetworkNodeTxError = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_errs_total{device!='lo'}[10s]))"

// KNetworkPodTxError: Returns the total number of errors transmitted by the pod over the past 1 minute.
const KNetworkPodTxError = "rate(container_network_transmit_errors_total{pod='%s'}[1m])"
