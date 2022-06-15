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


class PrometheusQuery4Edge(PrometheusClient):

    def __init__(self, port, ip):
        super().__init__(port, ip)

        self.container_id = None
        self.update_container_id()


    ###############################################################################
    # Utilitiy functions
    ###############################################################################
    @property
    def get_container_uptime(self):
        """
        List uptime of containers
        You will need this to find the container IDs
        Note: iot_executor is the name of image of executor deployment
        :return:
        """
        expr = "container_start_time_seconds{image=~'.*iot_executor'}"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_container_id(self):
        """
        Get container id
        :return:
        """
        return self.get_container_uptime['metric']['id']

    def update_container_id(self):
        """
        Get container id
        :return:
        """
        self.container_id = self.get_container_id
        return self.container_id

    ########################################################################################################################
    # CPU metrics
    ########################################################################################################################

    def get_node_cpu_user_usage_avg(self, interval='10s'):
        """
        Node CPU average over interval
        Warnings: the result may not available
        :param interval: default 10s
        :return:
        """
        expr = "rate(node_cpu_seconds_total{mode='user'}[%s])" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                cpu_counter = 0
                usage = 0.
                for cpu in res:
                    usage += float(cpu['value'][1])
                    cpu_counter +=1
                return float(usage / cpu_counter)

            time.sleep(1)
        return self.retry_fake_value

    def get_node_cpu_sys_usage_avg(self, interval='10s'):
        """
        Node CPU average over interval
        Warnings: the result may not available
        :param interval: default 10s
        :return:
        """
        expr = "rate(node_cpu_seconds_total{mode='system'}[%s])" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                cpu_counter = 0
                usage = 0.
                for cpu in res:
                    usage += float(cpu['value'][1])
                    cpu_counter +=1
                return float(usage / cpu_counter)

            time.sleep(1)
        return self.retry_fake_value

    def get_container_cpu_usage(self, interval='10s'):
        """
        Container CPU average over 10 seconds
        %s refers to the container ID. E.g: /docker/[container_id]
        Return all physical cpu, cpu0 cpu1....
        :param interval: default 10s
        :return:
        """
        expr = "rate(container_cpu_usage_seconds_total{id='%s'}[%s])" %(self.container_id,interval)
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
    def get_node_memory_usage(self):
        """
        Get Current memory usage
        :return:
        """
        expr = "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / (node_memory_MemTotal_bytes)"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_container_memory_usage(self):
        """
        Container memory usage
        %s refers to the container ID. E.g: /docker/[container_id]
        :return:
        """
        expr = "container_memory_usage_bytes{id='%s'}" %self.container_id
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_memory_peak(self, interval='30s'):
        """
        Peak node memory usage over the past interval
        :param interval: default 30s
        :return:
        """
        expr = "node_memory_MemTotal_bytes - max_over_time(node_memory_MemAvailable_bytes[%s])" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    def get_container_memory_peak(self, interval='10s'):
        """
        Peak node memory usage over the past interval
        :param interval: default 10s
        :return:
        """
        expr = "max_over_time(container_memory_usage_bytes{id='%s'}[%s])" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_container_memory_limit(self):
        """
        List soft memory limit of container
        Note: 0 means it is unset
        :return:
        """
        expr = "container_spec_memory_reservation_limit_bytes{id='%s'}" %self.container_id
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    @property
    def get_container_memory_limit_seconds(self, interval='5m', limit='1s'):
        """
        Seconds since containeer reached its soft memory limit
        %s is the container name
        :param interval: default 5m
        :param step: limit 1s
        :return:
        """
        expr = "count_over_time((container_memory_working_set_bytes{id='%s'} > " \
               "container_spec_memory_reservation_limit_bytes{id='%s'})[%s:%s])" %(self.container_id,
                                                                                   self.container_id,interval,limit)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    ########################################################################################################################
    # Disk metrics
    ########################################################################################################################

    def get_node_disk_read_bytes(self, interval='10s'):
        """
        Get node Read bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_disk_read_bytes_total[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[1]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_disk_write_bytes(self, interval='10s'):
        """
        Get node Write bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_disk_written_bytes_total[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[1]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_disk_read_bytes(self, interval='10s'):
        """
        Container read bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_fs_reads_bytes_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_disk_write_bytes(self, interval='10s'):
        """
        Container write bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_fs_writes_bytes_total{id='%s'}[%s]))" %(self.container_id,interval)
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
        :return:
        """
        expr = "(node_filesystem_size_bytes - node_filesystem_avail_bytes) / node_filesystem_size_bytes"
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    ########################################################################################################################
    # Network metrics
    ########################################################################################################################

    def get_node_network_receive_bytes(self, interval='10s'):
        """
        Network receive bytes over interval
        :param interval: default 10s
        :return:
        """
        # TODO Error
        expr = "max by (device, instance) (rate(node_network_receive_bytes_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_bytes(self, interval='10s'):
        """
        Network transmit bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_network_transmit_bytes_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_receive_packets(self, interval='10s'):
        """
        Network receive packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "rate(max by (device, instance) (node_network_receive_packets_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_packets(self, interval='10s'):
        """
        Network transmit packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "rate(max by (device, instance) (node_network_transmit_packets_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_receive_dropped_packets(self, interval='10s'):
        """
        NNetwork receive dropped packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_network_receive_drop_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_dropped_packets(self, interval='10s'):
        """
        Network transmit dropped packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_network_transmit_drop_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_receive_errors(self, interval='10s'):
        """
        Network receive errors over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_network_receive_errs_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_node_network_transmit_errors(self, interval='10s'):
        """
        Network transmit errors over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (device, instance) (rate(node_network_transmit_errs_total{device!='lo'}[%s]))" %interval
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_receive_bytes(self, interval='10s'):
        """
        Container network receive bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_receive_bytes_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_transmit_bytes(self, interval='10s'):
        """
        Container network transmit bytes over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_transmit_bytes_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return float(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_receive_packets(self, interval='10s'):
        """
        Container network receive packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_receive_packets_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_transmit_packets(self, interval='10s'):
        """
        Container network transmit packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_transmit_packets_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return res[0]
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_receive_dropped_packets(self, interval='10s'):
        """
        Container network receive dropped packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_receive_packets_dropped_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_transmit_dropped_packets(self, interval='10s'):
        """
        Container network transmit dropped packets over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_transmit_packets_dropped_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_receive_errors(self, interval='10s'):
        """
        Container network receive errors over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_receive_errors_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value

    def get_container_network_transmit_errors(self, interval='10s'):
        """
        Container network transmit errors over interval
        :param interval: default 10s
        :return:
        """
        expr = "max by (id) (rate(container_network_transmit_errors_total{id='%s'}[%s]))" %(self.container_id,interval)
        for _ in range(self.retry_times):
            res = self.extract_result(self.query_by_expr(expr))
            if res is not None:
                return int(res[0]['value'][1])
            time.sleep(1)
        return self.retry_fake_value
