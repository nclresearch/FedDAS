"""

    Warnings: None

    Author: Ringo Sham & Rui Sun
"""
import time
import sys,os

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from util_components.monitor.prometheus.prome_client import PrometheusClient
except ImportError:
    from UbiFL.util_components.monitor.prometheus.prome_client import PrometheusClient


class PrometheusQuery4Cloud(PrometheusClient):

    def __init__(self, ip, port):
        super().__init__(ip,port)

        self.hostname = "19scomps001"
        self.pod_expr = "pod=~'aggregator.*'"
        self.pod_name = None

        self.update_pod_name()

    @property
    def node_lookup(self):
        """
            This queries looks up the node exporter endpoint IP (which is random) by the node hostname (which is static)
            Format %s is the node hostname
            We will coine this with other queries to get the node metrics
        :return:
        """
        expr = "node_boot_time_seconds{kubernetes_namespace='ubifl_monitor'} * on(instance) group_right() label_replace(max by(pod_ip, node) (kube_pod_info{pod=~'node-exporter.*', node='%s'}), 'instance', '$1:9100', 'pod_ip', '(.*)')" %self.hostname
        return expr

    ###############################################################################
    # Utilitiy functions
    ###############################################################################

    @property
    def get_pod_info(self):
        """
        List all pods within the cluster
        We will need this to get the pod name (They are total random)
        :return:
        """
        expr = "kube_pod_info{%s}" %self.pod_expr

        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_pod_name(self):
        """
        Get actual name
        :return:
        """

        for _ in range(self.retry_times):
            res = self.get_pod_info
            if res['metric']['pod'] is not None:
                return res['metric']['pod']
            time.sleep(1)
        return self.retry_fake_value

    def update_pod_name(self):
        """
        Get actual name
        :return:
        """
        self.pod_name = self.get_pod_name
        return self.pod_name

    ########################################################################################################################
    # CPU metrics
    ########################################################################################################################

    def get_node_cpu_usage(self, interval='10s'):
        """
        Node CPU usage over interval
        :param interval: default 10s
        :return:
        """
        # TODO ERROR
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) sum by (instance, cpu) (rate(node_cpu_seconds_total{kubernetes_namespace='ubifl_monitor', mode!='idle'}[%s]))" % interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_cpu_usage(self, interval='30s'):
        """
        Pod CPU usage
        Warnings: Due to limitations from cAdvisor, we can only calculate the average of the pod CPU usage over the last minute
        :param interval:
        :return:
        """
        expr = "max by (pod) (rate(container_cpu_usage_seconds_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    ########################################################################################################################
    # Memory metrics
    ########################################################################################################################

    @property
    def get_node_mem_total_bytes(self):
        """
        Node memory usage currently
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) sum by (instance, node) (node_memory_MemTotal_bytes)"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_node_mem_usage(self):
        """
        Node memory usage currently
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) sum by (instance, node) (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes)"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                mem = self.get_node_mem_total_bytes
                if mem is None:
                    continue
                return float(res[0]['value'][1]) / mem

            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_node_mem_total_bytes(self):
        """
        Node memory usage currently
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) sum by (instance, node) (node_memory_MemTotal_bytes)"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_pod_mem_usage(self):
        """
        Pod memory usage currently
        :return:
        """
        expr = "max by (pod) (container_memory_usage_bytes{pod='%s'})" %self.pod_name
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                mem = self.get_node_mem_total_bytes
                if mem is None:
                    continue
                return float(res[0]['value'][1]) / mem
            time.sleep(1)
        return self.retry_fake_value

    def get_node_mem_usage_peak(self, interval='30s'):
        """
        Peak node memory usage over the past interval
        :param interval: default 30s
        :return:
        """
        #
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (instance, node) (node_memory_MemTotal_bytes - max_over_time(node_memory_MemAvailable_bytes[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]['value'][1]
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_mem_usage_peak(self, interval='30s'):
        """
        Peak pod memory usage over the past interval
        :param interval: default 30s
        :return:
        """
        # Peak pod memory usage over the past 5 minutes
        expr = "max(max_over_time(container_memory_usage_bytes{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]['value'][1]
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_pod_mem_usage_limit(self):
        """
        List soft memory limit of pod
        Note: 0 means it is unset
        :return:
        """
        expr = "container_spec_memory_reservation_limit_bytes{pod='%s'}" %self.pod_name
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]['value'][1]
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_mem_usage_limit_seconds(self, interval='5m', limit='1s'):
        """
        Seconds since pod reached its soft memory limit
        %s is the pod name
        :param interval: default 5m
        :param step: limit 1s
        :return:
        """
        expr = "count_over_time((container_memory_working_set_bytes{pod='%s'} > container_spec_memory_reservation_limit_bytes{pod='%s'})[%s:%s])" %(self.pod_name, self.pod_name,interval,limit)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]['value'][1]
            time.sleep(1)
        return self.retry_fake_value

    ########################################################################################################################
    # Disk metrics
    ########################################################################################################################

    def get_node_disk_read_bytes(self, interval='10s'):
        """
        Node read bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_disk_read_bytes_total[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_disk_write_bytes(self, interval='10s'):
        """
        Node write bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_disk_written_bytes_total[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_disk_read_bytes(self, interval='30s'):
        """
        Pod read bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (pod) (rate(container_fs_reads_bytes_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_disk_write_bytes(self, interval='30s'):
        """
        Pod write bytes overinterval
        :param interval: default 10s
        :return:
        """
        expr = "max by (pod) (rate(container_fs_writes_bytes_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_node_filesystem_usage(self):
        """
        Current filesystem usage in percentage
        Note: This lists out all mountpoints on the device, including tmpfs. And each mountpoint may use the same partition.
        :return: percentage
        """
        expr = self.node_lookup + "on(instance) group_right(node) max by (device, mountpoint, instance) ((node_filesystem_size_bytes - node_filesystem_avail_bytes) / node_filesystem_size_bytes)"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    ########################################################################################################################
    # Network metrics
    ########################################################################################################################

    def get_node_network_receive_bytes(self, interval='10s'):
        """
        Node network receive bytes over interval
        :param interval: default 10s
        :return: 
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_bytes_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_bytes(self, interval='10s'):
        """
        Node network transmit bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_bytes_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_receive_packets(self, interval='10s'):
        """
        Node network receive packets over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_packets_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_packets(self, interval='10s'):
        """
        Node network transmit packets over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_packets_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_receive_dropped_packets(self, interval='10s'):
        """
        Node network transmit packets over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_drop_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_dropped_packets(self, interval='10s'):
        """
        Node network transmit dropped packets over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_drop_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_receive_errors(self, interval='10s'):
        """
        Node network receive errors over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_receive_errs_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_errors(self, interval='10s'):
        """
        Node network transmit errors over interval
        :param interval: default 10s
        :return:
        """
        expr = self.node_lookup + "* 0 + on(instance) group_right(node) max by (device, instance) (rate(node_network_transmit_errs_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_receive_bytes(self, interval='30s'):
        """
        Pod network receive bytes over interval
        :param interval: default 30s
        :return: 
        """
        expr = "max by (pod) (rate(container_network_receive_bytes_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_transmit_bytes(self, interval='30s'):
        """
        Pod network transmit bytes over interval
        :param interval: default 30s
        :return: 
        """
        expr = "max by (pod) (rate(container_network_transmit_bytes_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_receive_packets(self, interval='30s'):
        """
        Pod network receive packets over interval
        :param interval: default 30s
        :return:
        """
        expr = "max by (pod) (rate(container_network_receive_packets_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_transmit_packets(self, interval='30s'):
        """
        Pod network transmit packets over interval
        :param interval: default 30s
        :return:
        """
        expr = "max by (pod) (rate(container_network_transmit_packets_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_receive_dropped_packets(self, interval='30s'):
        """
        Pod network receive dropped packets over interval
        :param interval: default 30s
        :return:
        """
        expr = "max by (pod) (rate(container_network_receive_packets_dropped_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_transmit_dropped_packets(self, interval='30s'):
        """
        Pod network transmit dropped packets over interval
        :param interval: default 30s
        :return:
        """
        expr = "max by (pod) (rate(container_network_transmit_packets_dropped_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_receive_errors(self, interval='30s'):
        """
        Pod network receive errors over interval
        :param interval: default 30s
        :return:
        """
        expr = "max by (pod) (rate(container_network_receive_errors_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_pod_network_transmit_errors(self, interval='30s'):
        """
        Pod network transmit errors over interval
        :param interval: default 30s
        :return:
        """
        expr = "max by (pod) (rate(container_network_transmit_errors_total{pod='%s'}[%s]))" %(self.pod_name,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value