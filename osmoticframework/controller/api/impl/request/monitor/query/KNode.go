package query

//Utility queries

//KEndpointInfo: Displays a map which lists out Node exporter's endpoint IP and the node's name that Node exporter is deployed into
//Node exporter is deployed once on every node
const KEndpointInfo = "node_boot_time_seconds{kubernetes_namespace='monitoring'} * on(instance) group_right() label_replace(max by(pod_ip, node) (kube_pod_info{pod=~'node-exporter.*', node='%s'}), 'instance', '$1:9100', 'pod_ip', '(.*)')"

//All node level queries have the endpoint IP labelled. This refers to the endpoint IP of node exporter.
//Since node exporter is deployed once on every node in the cluster, we can make use of Prometheus queries to look up endpoint IPs based on the name of the node
//Then combine it with the queries that query metrics to achieve something along the line:
//Input: Name of the node
//Output: {endpoint_ip, cpu_core, node_name : metric_value}

//KNodeQuery : Allows users to look up the endpoint IP of node exporter based on the name of the node
//This returns the following: {endpoint_ip, node_name, pod_ip : value}
//The value in here does not matter. We can nullify it by multiply it by 0
//This query is then combined with other queries to get metrics based on node name alone
//node_boot_time_seconds{kubernetes_namespace='monitoring'} * on(instance) group_right() label_replace(max by(pod_ip, node) (kube_pod_info{pod=~'node-exporter.*', node='cloud-server'}), 'instance', '$1:9100', 'pod_ip', '(.*)') * 0 +
const KNodeQuery = KEndpointInfo + " * 0 + "
