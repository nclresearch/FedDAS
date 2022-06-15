package query

//CPU metric queries for Kubernetes

//KCPUNodeAvg : This first queries all endpoint IPs in the cluster. Then filter out results that are not related to node exporter. Next it merges the result of CPU average metric
//The result metric would be: {cpu_core_num, endpoint_ip, node_name : cpu_usage_per_core_over_10_sec}
const KCPUNodeAvg = KNodeQuery + "on(instance) group_right(node) sum by (instance, cpu) (rate(node_cpu_seconds_total{kubernetes_namespace='monitoring', mode!='idle'}[10s]))"

//KCPUPodAvg : Queries CPU usage by pod over the past 1 minute
//It must be 1 minute as cAdvisor isn't collecting CPU usage frequently
const KCPUPodAvg = "max by (pod) (rate(container_cpu_usage_seconds_total{pod='%s'}[1m]))"

//KCPUTime : Queries total CPU time (CPUTime spent on CPU load) on node level
const KCPUTime = KNodeQuery + "on(instance) group_right(node) sum by (cpu, instance) (node_cpu_seconds_total{mode!='idle'})"

//KCPUUtil : Queries total CPU usage over the past 10 seconds on node level
const KCPUUtil = KNodeQuery + "on(instance) group_right(node) (sum by (instance) (sum by (cpu, instance) (irate(node_cpu_seconds_total{mode!='idle'}[10s]))) / (count by (instance) (count by (cpu, instance) (node_cpu_seconds_total))))"
