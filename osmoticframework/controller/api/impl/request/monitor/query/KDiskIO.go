package query

//Disk IO metric queries for Kubernetes

//KIONodeTime : Returns of how much time is spent on IO operation on node level over the past 10 seconds
//If the time returned is high, either the system is thrashing or there is a hard drive failure
const KIONodeTime = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_disk_io_time_seconds_total[10s]))"

//KIOPodTime : Returns how much time is spent on IO operation on pod level over the past 1 minute
const KIOPodTime = "max by (device, pod) (rate(container_fs_io_time_seconds_total{pod='%s'}[1m]))"

//KIOReadNodeBytes : Returns how many bytes read per disk on the node level over the past 10 seconds
const KIOReadNodeBytes = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_disk_read_bytes_total[10s]))"

//KIOReadPodBytes : Returns how many bytes read per disk on the pod level over the past 1 minute
const KIOReadPodBytes = "max by (device, pod) (rate(container_fs_reads_bytes_total{pod='%s'}[1m]))"

//KIOWriteNodeBytes : Returns how many bytes written per disk on the node level over the past 10 seconds
const KIOWriteNodeBytes = KNodeQuery + "on(instance) group_right(node) max by (device, instance) (rate(node_disk_write_bytes_total[10s]))"

//KIOWritePodBytes : Returns how many bytes written per disk on the pod level over the past 1 minute
const KIOWritePodBytes = "max by (device, pod) (rate(container_fs_writes_bytes_total{pod='%s'}[1m]))"

//KIOFilesystemUsedBytes : Returns how many bytes used per disk on the node level
const KIOFilesystemUsedBytes = KNodeQuery + "on(instance) group_right(node) max by (device, mountpoint, instance) (node_filesystem_size_bytes - node_filesystem_avail_bytes)"

//KIOFilesystemSizeBytes : Returns how many bytes available per disk on the node level
const KIOFilesystemSizeBytes = KNodeQuery + "on(instance) group_right(node) max by (device, mountpoint, instance) (node_filesystem_size_bytes)"
