package query

//Memory metric queries for Kubernetes

// KMemoryPod: Returns current memory usage on a pod
const KMemoryPod = "max by (pod) (container_memory_usage_bytes{pod='%s'})"

// KMemoryNode: Returns current memory usage on a node (Total - available)
const KMemoryNode = KNodeQuery + "on(instance) group_right(node) max by (instance, node) (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes)"

// KMemoryPodPeak: Returns peak memory usage on a pod over the past 5 minutes
const KMemoryPodPeak = "max(max_over_time(container_memory_working_set_bytes{pod='%s'}[5m]))"

// KMemoryNodePeak: Returns peak memory usage on a node over the past 5 minutes
const KMemoryNodePeak = KNodeQuery + "on(instance) group_right(node) max by (instance, node) (node_memory_MemTotal_bytes - max_over_time(node_memory_MemAvailable_bytes[5m]))"

// KSoftLimitQuery: Returns soft memory limit on a pod in bytes
const KSoftLimitQuery = "container_spec_memory_reservation_limit_bytes{pod='%s'}"

// KMemoryPodReachLimitSeconds: Returns the number of seconds since the pod reached its soft memory limit
const KMemoryPodReachLimitSeconds = "count_over_time((container_memory_working_set_bytes{pod='%[1]s'} > container_spec_memory_reservation_limit_bytes{pod='%[1]s'})[5m:1s])"
