"""
    Client

    Warnings: None

    Author: Rui Sun
"""
import pickle
import threading
import time

import torch.utils.data
import wandb
import sys, os

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from api.utils.optimizers import Optimizers
    from core.utils.client_libs import *
    from util_components.monitor.model import GPUEventTimer
    from util_components.monitor.prometheus.prome_query_4_edge import PrometheusQuery4Edge
    from DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient
except ImportError:
    from UbiFL.api.utils.optimizers import Optimizers
    from UbiFL.core.utils.client_libs import *
    from UbiFL.util_components.monitor.model import GPUEventTimer
    from UbiFL.util_components.monitor.prometheus.prome_query_4_edge import PrometheusQuery4Edge
    from UbiFL.DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient

from torch.utils.data import DataLoader
from tqdm import tqdm

from torch import nn
import requests


class FLClient:
    def __init__(self, hostId, clientId, publish_topic, dev):

        # ======== Environment ========
        self.test_dataset = None
        self.train_dataset = None
        self.hostId = hostId
        self.clientId = clientId
        self.dev = dev
        self.all_test_accuracy = []
        self.all_test_loss = []
        self.all_global_accuracy = []
        self.all_global_loss = []
        self.all_train_accuracy = []
        self.all_train_loss = []

        # ======== Communication ========
        self.publish_topic = publish_topic
        self.total_communication_round = 0

        # ======== Data ========
        self.train_dataloader = None
        self.test_dataloader = None

        # ======== System Utility ========
        self.score = 0

        # ======== Runtime Information ========
        self.current_communication_round = 1
        self.loss_func = nn.CrossEntropyLoss()

        self.is_under_training = False

        # ======== Monitoring ========
        # InfluxDB client
        if args.is_monitoring:
            self.influx_client = FLInfluxDBClient(token=args.influxdb_token, org=args.influxdb_org, ip=args.influxdb_ip,
                                                  port=args.influxdb_port, timeout=1_000_000)
            self.influx_client.connect()
            self.exec_monitor = ExecutorMonitor(hostId=self.hostId, clientId=self.clientId)
            self.exec_monitor_thread = threading.Thread(target=self.monitoring)
            # self.monitoring_locker = threading.Lock()
            # self.prome_query = PrometheusQuery4Edge(args.prometheus_ip, args.prometheus_port)
            self.exec_monitor_thread.start()



            # ============= Record Hyper-parameters =============
            sequence = [
                f"hyper_parameters,device_id={self.hostId},client_id={self.clientId} learning_rate={args.learning_rate}",
                f"hyper_parameters,device_id={self.hostId},client_id={self.clientId} local_epoch={args.local_epoch}",
                f"hyper_parameters,device_id={self.hostId},client_id={self.clientId} batch_size={args.batch_size}",
                f"hyper_parameters,device_id={self.hostId},client_id={self.clientId} test_batch_size={args.test_batch_size}"
            ]
            self.influx_client.write_new_records(bucket="pre_setting", data=sequence)

    def init_model(self, global_params):
        # init model
        if args.model_from == "basic":
            self.local_model = get_basic_model(task=args.task, model_name=args.model_name).to(self.dev)
            # Update local model
        elif args.model_from == "nas":
            self.local_model = get_model_from_nas(task=args.task, model_stru=global_params['model_structure'])

        self.local_parameters = {}
        for key, var in self.local_model.state_dict().items():
            self.local_parameters[key] = var.clone()

        if args.optimizer == "adam":
            self.optimizer = Optimizers(self.local_model.parameters(), float(args.learning_rate)).get_adam_optimizer()
        elif args.optimizer == "sgd":
            self.optimizer = Optimizers(self.local_model.parameters(), float(args.learning_rate)).get_sgd_optimizer()

    # def getScore(self):
    #     return self.score

    def localUpdate(self, localEpoch, localBatchSize, lossFun, opti, train_dataset, test_dataset):

        self.local_model = self.local_model.to(self.dev)

        # As there already sampled, so set shuffle=False
        self.train_dataloader = DataLoader(train_dataset, batch_size=localBatchSize, shuffle=True)
        self.test_dataloader = DataLoader(test_dataset, batch_size=localBatchSize, shuffle=True)

        train_loss = 0.
        total_size = 0

        # ============ Before train global model eval ============
        # with torch.no_grad():
        #     global_acc, global_loss,_ = self.__local_test(self.local_model)
        # <<<<<<<<<< Monitoring <<<<<<<<<<

        counter = 0
        total_correct = 0

        forward_propagation_intervals = []
        backward__propagation_intervals = []
        all_epoch_latency = []

        for epoch in tqdm(range(localEpoch), desc=f"Client:{self.clientId} under training..."):

            if args.is_monitoring:
                # Millisecond
                t0 = int(round(time.time() * 1000))

            self.local_model.train()
            for data, label in self.train_dataloader:
                counter += 1
                data, label = data.to(self.dev), label.to(self.dev)
                opti.zero_grad()

                # init monitor timer for forward propagation
                if args.is_monitoring:
                    if self.dev == "cuda":
                        f_start, f_end = GPUEventTimer(enable_timing=True), GPUEventTimer(enable_timing=True)
                        f_start.record()
                    else:
                        f_start = time.time()
                output = self.local_model(data)
                if args.is_monitoring:
                    # ============= Epoch delay monitoring =============
                    all_epoch_latency.append(int(round(time.time() * 1000)) - t0)

                    # ============= Model performance delay monitoring =============
                    if self.dev == "cuda":
                        # Wait for all task finish
                        f_end.synchronize()
                        f_end.record()
                        forward_propagation_intervals.append(f_start.elapsed_time(f_end))
                    else:
                        # ms
                        forward_propagation_intervals.append((time.time() - f_start) * 1000)

                # the output type of network generated from NAS is type
                if type(output) is tuple:
                    output = output[0]

                # val
                predicted = torch.argmax(output, dim=1)
                total_correct += predicted.eq(label.view_as(predicted)).sum().item()

                loss = lossFun(output, label)

                # init monitor timer for backward propagation
                if args.is_monitoring:
                    if self.dev == "cuda":
                        b_start, b_end = GPUEventTimer(enable_timing=True), GPUEventTimer(enable_timing=True)
                        b_start.record()
                    else:
                        b_start = time.time()
                loss.backward()
                if args.is_monitoring:
                    if self.dev == "cuda":
                        # Wait for all task finish
                        b_end.synchronize()
                        b_end.record()
                        backward__propagation_intervals.append(b_start.elapsed_time(b_end))
                    else:
                        # ms
                        backward__propagation_intervals.append(round((time.time() - b_start) * 1000, 3))

                train_loss += loss.item()
                opti.step()
                total_size += label.size(0)

        train_loss = round(train_loss / counter, 4)
        train_acc = total_correct / total_size
        train_acc *= 100.

        if self.current_communication_round % args.local_test_freq == 0:
            # Local testing
            with torch.no_grad():
                acc, loss, inf_latency = self.__local_test(self.local_model)
                loss = round(loss, 4)
                self.all_test_accuracy.append(acc)
                self.all_test_loss.append(loss)
                # global_loss = round(global_loss, 4)
                #
                # self.all_global_accuracy.append(global_acc)
                # self.all_global_loss.append(global_loss)
                self.all_train_accuracy.append(train_acc)
                self.all_train_loss.append(train_loss)

                wandb.log({
                    "Test/Acc": acc,
                    "Test/Loss": loss,
                    "Train/Acc": train_acc,
                    "Train/Loss": train_loss,
                    # "global_model_test/Acc": global_acc,
                    # "global_model_test/Loss": global_loss,
                })

                logger.info(
                    f'{self.clientId}, local model train accuracy: {train_acc}, local model train loss: {train_loss}')
                logger.info(f'{self.clientId}, local model test accuracy: {acc}, local model test loss: {loss}')

                # >>>>>>>>>> Monitoring >>>>>>>>>>
                if args.is_monitoring:

                    # Avoid GPU warm up impact , random value 10
                    if len(forward_propagation_intervals) > 10:
                        forward_propagation_intervals = forward_propagation_intervals[10:]
                        backward__propagation_intervals = backward__propagation_intervals[10:]
                    forward_propagation_intervals_avg = round(
                        sum(forward_propagation_intervals) / len(forward_propagation_intervals), 3)
                    backward_propagation_intervals_avg = round(
                        sum(backward__propagation_intervals) / len(backward__propagation_intervals), 3)

                    # logger.info(f'{self.clientId}, global model accuracy: {global_acc}, global model loss: {global_loss}')
                    logger.info(
                        f'{self.clientId}, model inference latency: {inf_latency}, forward propagation latency in avg: '
                        f'{forward_propagation_intervals_avg}, backward propagation latency in avg: {backward_propagation_intervals_avg}')

                    wandb.log({
                        "model_monitor/Inf_latency": inf_latency,
                        "model_monitor/Forward_propagation_latency_avg": forward_propagation_intervals_avg,
                        "model_monitor/Backward_propagation_latency_avg": backward_propagation_intervals_avg,
                    })

                    sequence = [
                        f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} epoch_delay_avg={np.mean(all_epoch_latency)}",
                        f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} inf_latency_avg={inf_latency}",
                        f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} forward_propagation_intervals_avg={forward_propagation_intervals_avg}",
                        f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} backward_propagation_intervals_avg={backward_propagation_intervals_avg}",
                        f"local_model_result,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} test_acc={acc}",
                        f"local_model_result,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} test_loss={loss}",
                        f"local_model_result,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} train_acc={train_acc}",
                        f"local_model_result,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} train_loss={train_loss}"
                        # f"global_model_result,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} global_test_acc={global_acc}",
                        # f"global_model_result,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} global_test_loss={global_loss}"
                    ]
                    # sequence = data + self.prometheus_monitor()
                    # logger.info(sequence)
                    self.influx_client.write_new_records(bucket="running_monitoring", data=sequence)

                # <<<<<<<<<< Monitoring <<<<<<<<<<

        if self.current_communication_round == self.total_communication_round:
            torch.save(self.all_test_accuracy, 'cache/accuracy/test-' + self.clientId + '.pkl')
            torch.save(self.all_test_loss, 'cache/loss/test-' + self.clientId + '.pkl')
            # torch.save(self.all_global_accuracy, 'cache/accuracy/global-' + self.clientId + '.pkl')
            # torch.save(self.all_global_loss, 'cache/loss/global-' + self.clientId + '.pkl')
            torch.save(self.all_train_accuracy, 'cache/accuracy/train-' + self.clientId + '.pkl')
            torch.save(self.all_test_accuracy, 'cache/loss/train-' + self.clientId + '.pkl')

            self.is_under_training = False
            self.all_test_accuracy.clear()
            self.all_test_accuracy.clear()
            # self.all_global_accuracy.clear()
            # self.all_global_loss.clear()
            self.all_train_accuracy.clear()
            self.all_train_loss.clear()

            if args.is_monitoring:
                self.influx_client.close()

            self.exec_monitor.close()

        # If the aggregator server is cpu based
        # return Net.to("cpu").state_dict()

    def request_dataset(self, dir, path):
        url = 'http://' + args.ps_ip + ':' + str(args.parameter_server_client_register_port) \
              + args.parameter_server_api_download_dataset
        # logger.debug(url)
        params = {'dir': dir, 'path': path}
        # headers = {'Connection': 'keep-alive', 'Accept-Encoding': 'gzip, deflate',
        #            'Accept': '*/*', 'User-Agent': 'python-requests/2.18.3'}

        resp = requests.get(url=url, params=params)

        # save file
        with open(dir + path, 'wb') as f:
            f.write(resp.content)
            f.close()

    def on_message(self, client, userdata, msg):
        # Callback func

        logger.debug(f'Client {self.clientId} got a message from server')
        server_payload = pickle.loads(msg.payload)

        if args.is_monitoring:
            # ============= Communication delay monitoring (Network) =============
            delta_t = int(round(time.time() * 1000)) - int(server_payload["monitoring"]["sent_time"])

            data = [
                f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={self.current_communication_round} comm_delay={delta_t}"]
            # sequence = data + self.prometheus_monitor()

            self.influx_client.write_new_records(bucket="running_monitoring", data=data)

        global_model_param = server_payload['global_model']

        # Update comm
        self.total_communication_round = int(server_payload['communication']['total_communication_round'])
        self.current_communication_round = int(server_payload['communication']['current_communication_round'])

        # Init model
        if self.current_communication_round == 1:
            self.init_model(global_model_param)
            self.is_under_training = True

        # Update local model
        self.local_model.load_state_dict(global_model_param['params'], strict=True)
        logger.info(f' {self.clientId} finished load global model')

        # Client local training
        file_patch_dict = {
            "train_dataset_path": args.exp_path + "/cache/train_data/partition/",
            "train_data_file_name": "train-" + self.clientId + ".pth",
            "test_dataset_path": args.exp_path + "/cache/test_data/partition/",
            "test_data_file_name": "test-" + self.clientId + ".pth",
        }

        # Download dataset from server
        if not os.path.exists(file_patch_dict["train_dataset_path"] + file_patch_dict["train_data_file_name"]):
            logger.info(f"{self.clientId} start to download train data ...")
            # train data
            self.request_dataset(dir=file_patch_dict["train_dataset_path"],
                                 path=file_patch_dict["train_data_file_name"])

        if not os.path.exists(file_patch_dict["test_dataset_path"] + file_patch_dict["test_data_file_name"]):
            logger.info(f"{self.clientId} start to download test data ...")
            # test data
            self.request_dataset(dir=file_patch_dict["test_dataset_path"],
                                 path=file_patch_dict["test_data_file_name"])

        if self.current_communication_round == 1:
            logger.info(f"{self.clientId} start to load training data ...")
            self.train_dataset = torch.load(
                file_patch_dict["train_dataset_path"] + file_patch_dict["train_data_file_name"])
            logger.info(f"{self.clientId} finished to load train data")

            logger.info(f"{self.clientId} start to load test data ...")
            self.test_dataset = torch.load(
                file_patch_dict["test_dataset_path"] + file_patch_dict["test_data_file_name"])
            logger.info(f"{self.clientId} finished to load test data")

        self.localUpdate(args.local_epoch, args.batch_size, self.loss_func, self.optimizer, self.train_dataset,
                         self.test_dataset)

        # Update local model param
        # To target server computing unite incase the aggregator server is cpu based
        # If the aggregator server is cpu based
        local_model_params_dict = self.local_model.to(server_payload['sys']['dev']).state_dict()

        for key, var in local_model_params_dict.items():
            self.local_parameters[key] = var.clone()

        # 组建回传参数
        payload_to_server = {
            "round": self.current_communication_round,
            "clientId": self.clientId,
            "parameters": {
                "weight_parameters": {
                    "local": self.local_parameters
                }
            }
        }

        if args.is_monitoring:
            payload_to_server["monitoring"] = {
                "sent_time": str(int(round(time.time() * 1000)))
            }

        # Publish 回server
        client.publish(topic=self.publish_topic, qos=2, payload=payload_to_server)

        # if int(server_payload['communication']['current_communication_round']) == int(server_payload['communication']['total_communication_round']):
        #     self.is_under_training = False

    def __local_test(self, model):
        model.eval()
        total_correct = 0.
        total_test_loss = 0.
        counter = 0
        total_size = 0
        inf_latencies = []

        with torch.no_grad():
            for data, label in self.test_dataloader:
                counter += 1
                data, label = data.to(self.dev), label.to(self.dev)
                # init monitor timer for backward propagation
                if args.is_monitoring:
                    if self.dev == "cuda":
                        f_start, f_end = GPUEventTimer(enable_timing=True), GPUEventTimer(enable_timing=True)
                        f_start.record()
                    else:
                        f_start = time.time()

                preds = model(data)

                if args.is_monitoring:
                    if self.dev == "cuda":
                        # Wait for all task finish
                        f_end.synchronize()
                        f_end.record()
                        inf_latencies.append(f_start.elapsed_time(f_end))
                    else:
                        # ms
                        inf_latencies.append((time.time() - f_start) * 1000)

                # the output type of network generated from NAS is type
                if type(preds) is tuple:
                    preds = preds[0]
                total_test_loss += self.loss_func(preds, label).item()
                predicted = torch.argmax(preds, dim=1)
                total_correct += predicted.eq(label.view_as(predicted)).sum().item()
                total_size += label.size(0)
            # total_size = len(self.test_dataloader)
            total_test_loss /= counter
            total_correct /= total_size

            if args.is_monitoring:
                # Note: Avoid GPU warm up
                if len(inf_latencies) > 10:
                    inf_latencies = inf_latencies[10:]

                inf_latency_avg = round(sum(inf_latencies) / len(inf_latencies), 3)
                return total_correct * 100., total_test_loss, inf_latency_avg

        return total_correct * 100., total_test_loss, None

    def monitoring(self):
        while True:
            if self.is_under_training:
                sequence = self.exec_monitor.prometheus_monitor(self.current_communication_round)
                # logger.info(sequence)
                self.exec_monitor.record_metrics(sequence)
            time.sleep(5)


'''
    Monitoring
'''


class ExecutorMonitor(object):
    def __init__(self, hostId, clientId):
        self.hostId = hostId
        self.clientId = clientId

        self.influx_client = FLInfluxDBClient(token=args.influxdb_token, org=args.influxdb_org, ip=args.influxdb_ip,
                                              port=args.influxdb_port, timeout=1_000_000)
        self.influx_client.connect()
        self.prome_query = PrometheusQuery4Edge(args.prometheus_ip, args.prometheus_port)

        try:
            from util_components.monitor.jetson_jtop.jtop import JTopMonitor

        except ImportError:
            from UbiFL.util_components.monitor.jetson_jtop.jtop import JTopMonitor

        self.jetson = JTopMonitor()
        self.jetson.start()


    def prometheus_monitor(self, current_communication_round):
        if self.prome_query.container_id is None:
            self.prome_query.update_container_id()

        # self.monitoring_locker.acquire()
        metrics = self.get_metrics()

        monitoring_list = [
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_cpu_usage={-1 if metrics['get_container_cpu_usage'] is None else round(metrics['get_container_cpu_usage'], 4)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_memory_usage={-1 if metrics['get_container_memory_usage'] is None else round(metrics['get_container_memory_usage'] / 1000 / 1000, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_disk_read_bytes={-1 if metrics['get_container_disk_read_bytes'] is None else round(metrics['get_container_disk_read_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_disk_write_bytes={-1 if metrics['get_container_disk_write_bytes'] is None else round(metrics['get_container_disk_write_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_network_receive_bytes={-1 if metrics['get_container_network_receive_bytes'] is None else round(metrics['get_container_network_receive_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_network_transmit_bytes={-1 if metrics['get_container_network_transmit_bytes'] is None else round(metrics['get_container_network_transmit_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_network_receive_errors={-1 if metrics['get_container_network_receive_errors'] is None else metrics['get_container_network_receive_errors']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} container_network_transmit_errors={-1 if metrics['get_container_network_transmit_errors'] is None else metrics['get_container_network_transmit_errors']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_network_receive_bytes={-1 if metrics['get_node_network_receive_bytes'] is None else round(metrics['get_node_network_receive_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_network_transmit_bytes={-1 if metrics['get_node_network_transmit_bytes'] is None else round(metrics['get_node_network_transmit_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_network_receive_errors={-1 if metrics['get_node_network_receive_errors'] is None else metrics['get_node_network_receive_errors']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_network_transmit_errors={-1 if metrics['get_node_network_transmit_errors'] is None else metrics['get_node_network_transmit_errors']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_cpu_user_usage_avg={-1 if metrics['get_node_cpu_user_usage_avg'] is None else round(metrics['get_node_cpu_user_usage_avg'], 4)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_cpu_sys_usage_avg={-1 if metrics['get_node_cpu_sys_usage_avg'] is None else round(metrics['get_node_cpu_sys_usage_avg'], 4)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_memory_usage={-1 if metrics['get_node_memory_usage'] is None else round(metrics['get_node_memory_usage'], 4)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_disk_read_bytes={-1 if metrics['get_node_disk_read_bytes'] is None else round(metrics['get_node_disk_read_bytes'] / 30, 2)}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} node_disk_write_bytes={-1 if metrics['get_node_disk_write_bytes'] is None else round(metrics['get_node_disk_write_bytes'] / 30, 2)}",

            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jetson_stats_cpu1={metrics['get_jtop_stats_cpu1']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jetson_stats_cpu2={metrics['get_jtop_stats_cpu2']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jetson_stats_cpu3={metrics['get_jtop_stats_cpu3']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jetson_stats_cpu4={metrics['get_jtop_stats_cpu4']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_gpu={metrics['get_jtop_stats_gpu']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_ram={metrics['get_jtop_stats_ram']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_emc={metrics['get_jtop_stats_emc']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_iram={metrics['get_jtop_stats_iram']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_swap={metrics['get_jtop_stats_swap']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_ape={metrics['get_jtop_stats_ape']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_temperature_a0={metrics['get_jtop_stats_temperature_a0']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_temperature_cpu={metrics['get_jtop_stats_temperature_cpu']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_temperature_gpu={metrics['get_jtop_stats_temperature_gpu']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_temperature_pll={metrics['get_jtop_stats_temperature_pll']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_temperature_thermal={metrics['get_jtop_stats_temperature_thermal']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_power_cur={metrics['get_jtop_stats_power_cur']}",
            f"client_running_status,device_id={self.hostId},client_id={self.clientId},comm={current_communication_round} jtop_stats_power_avg={metrics['get_jtop_stats_power_avg']}"
        ]
        # self.monitoring_locker.release()
        # logger.info(monitoring_list)
        return monitoring_list

    def get_metrics(self):

        jtop_stats = self.jetson.stats
        metrics = {
            "get_container_cpu_usage": self.prome_query.get_container_cpu_usage('30s'),
            "get_container_memory_usage": self.prome_query.get_container_memory_usage,
            "get_container_disk_read_bytes": self.prome_query.get_container_disk_read_bytes('30s'),
            "get_container_disk_write_bytes": self.prome_query.get_container_disk_write_bytes('30s'),
            "get_container_network_receive_bytes": self.prome_query.get_container_network_receive_bytes('30s'),
            "get_container_network_transmit_bytes": self.prome_query.get_container_network_transmit_bytes('30s'),
            "get_container_network_receive_errors": self.prome_query.get_container_network_receive_errors('30s'),
            "get_container_network_transmit_errors": self.prome_query.get_container_network_transmit_errors('30s'),
            "get_node_network_receive_bytes": self.prome_query.get_node_network_receive_bytes('30s'),
            "get_node_network_transmit_bytes": self.prome_query.get_node_network_transmit_bytes('30s'),
            "get_node_network_receive_errors": self.prome_query.get_node_network_receive_errors('30s'),
            "get_node_network_transmit_errors": self.prome_query.get_node_network_transmit_errors('30s'),
            "get_node_cpu_user_usage_avg": self.prome_query.get_node_cpu_user_usage_avg('30s'),
            "get_node_memory_usage": self.prome_query.get_node_memory_usage,
            "get_node_cpu_sys_usage_avg": self.prome_query.get_node_cpu_sys_usage_avg('30s'),
            "get_node_disk_read_bytes": self.prome_query.get_node_disk_read_bytes('30s'),
            "get_node_disk_write_bytes": self.prome_query.get_node_disk_write_bytes('30s'),

            # Result from jtop
            "get_jtop_stats_cpu1" : jtop_stats["CPU1"],
            "get_jtop_stats_cpu2": jtop_stats["CPU2"],
            "get_jtop_stats_cpu3": jtop_stats["CPU3"],
            "get_jtop_stats_cpu4": jtop_stats["CPU4"],
            "get_jtop_stats_gpu": jtop_stats["GPU"],
            "get_jtop_stats_ram": jtop_stats["RAM"],
            "get_jtop_stats_emc": jtop_stats["EMC"],
            "get_jtop_stats_iram": jtop_stats["IRAM"],
            "get_jtop_stats_swap": jtop_stats["SWAP"],
            "get_jtop_stats_ape": jtop_stats["APE"],
            # fan hardware is not available now
            # "get_jtop_": jtop_stats["fan"],
            "get_jtop_stats_temperature_a0": jtop_stats["Temp AO"],
            "get_jtop_stats_temperature_cpu": jtop_stats["Temp CPU"],
            "get_jtop_stats_temperature_gpu": jtop_stats["Temp GPU"],
            "get_jtop_stats_temperature_pll": jtop_stats["Temp PLL"],
            # wifi is not available now
            # "get_jtop_": jtop_stats["Temp iwlwifi"],
            "get_jtop_stats_temperature_thermal": jtop_stats["Temp thermal"],
            "get_jtop_stats_power_cur": jtop_stats["power cur"],
            "get_jtop_stats_power_avg": jtop_stats["power avg"],
        }

        return metrics

    def record_metrics(self, sequence):
        self.influx_client.write_new_records(bucket="running_monitoring", data=sequence)

    def close(self):
        self.influx_client.close()
        self.jetson.close()

if __name__ == '__main__':

    import time, os, sys

    sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
    sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
    try:
        from util_components.monitor.prometheus.prome_query_4_edge import PrometheusQuery4Edge
        from DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient
    except ImportError:
        from UbiFL.util_components.monitor.prometheus.prome_query_4_edge import PrometheusQuery4Edge
        from UbiFL.DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient

    prome_query = PrometheusQuery4Edge("172.17.0.1", "9090")

    while True:
        metrics = {
            "get_container_cpu_usage": prome_query.get_container_cpu_usage('30s'),
            "get_container_memory_usage": prome_query.get_container_memory_usage,
            "get_container_disk_read_bytes": prome_query.get_container_disk_read_bytes('30s'),
            "get_container_disk_write_bytes": prome_query.get_container_disk_write_bytes('30s'),
            "get_container_network_receive_bytes": prome_query.get_container_network_receive_bytes('30s'),
            "get_container_network_transmit_bytes": prome_query.get_container_network_transmit_bytes('30s'),
            "get_container_network_receive_errors": prome_query.get_container_network_receive_errors('30s'),
            "get_container_network_transmit_errors": prome_query.get_container_network_transmit_errors('30s'),
            "get_node_network_receive_bytes": prome_query.get_node_network_receive_bytes('30s'),
            "get_node_network_transmit_bytes": prome_query.get_node_network_transmit_bytes('30s'),
            "get_node_network_receive_errors": prome_query.get_node_network_receive_errors('30s'),
            "get_node_network_transmit_errors": prome_query.get_node_network_transmit_errors('30s'),
            "get_node_cpu_user_usage_avg": prome_query.get_node_cpu_user_usage_avg('30s'),
            "get_node_memory_usage": prome_query.get_node_memory_usage,
            "get_node_cpu_sys_usage_avg": prome_query.get_node_cpu_sys_usage_avg('30s'),
            "get_node_disk_read_bytes": prome_query.get_node_disk_read_bytes('30s'),
            "get_node_disk_write_bytes": prome_query.get_node_disk_write_bytes('30s')
        }

        monitoring_list = [
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_cpu_usage={-1 if metrics['get_container_cpu_usage'] is None else round(metrics['get_container_cpu_usage'], 4)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_memory_usage={-1 if metrics['get_container_memory_usage'] is None else round(metrics['get_container_memory_usage'] / 1000 / 1000, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_disk_read_bytes={-1 if metrics['get_container_disk_read_bytes'] is None else round(metrics['get_container_disk_read_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_disk_write_bytes={-1 if metrics['get_container_disk_write_bytes'] is None else round(metrics['get_container_disk_write_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_network_receive_bytes={-1 if metrics['get_container_network_receive_bytes'] is None else round(metrics['get_container_network_receive_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_network_transmit_bytes={-1 if metrics['get_container_network_transmit_bytes'] is None else round(metrics['get_container_network_transmit_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_network_receive_errors={-1 if metrics['get_container_network_receive_errors'] is None else metrics['get_container_network_receive_errors']}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 container_network_transmit_errors={-1 if metrics['get_container_network_transmit_errors'] is None else metrics['get_container_network_transmit_errors']}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_network_receive_bytes={-1 if metrics['get_node_network_receive_bytes'] is None else round(metrics['get_node_network_receive_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_network_transmit_bytes={-1 if metrics['get_node_network_transmit_bytes'] is None else round(metrics['get_node_network_transmit_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_network_receive_errors={-1 if metrics['get_node_network_receive_errors'] is None else metrics['get_node_network_receive_errors']}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_network_transmit_errors={-1 if metrics['get_node_network_transmit_errors'] is None else metrics['get_node_network_transmit_errors']}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_cpu_user_usage_avg={-1 if metrics['get_node_cpu_user_usage_avg'] is None else round(metrics['get_node_cpu_user_usage_avg'], 4)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_cpu_sys_usage_avg={-1 if metrics['get_node_cpu_sys_usage_avg'] is None else round(metrics['get_node_cpu_sys_usage_avg'], 4)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_memory_usage={-1 if metrics['get_node_memory_usage'] is None else round(metrics['get_node_memory_usage'], 4)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_disk_read_bytes={-1 if metrics['get_node_disk_read_bytes'] is None else round(metrics['get_node_disk_read_bytes'] / 30, 2)}",
            f"client_running_status,device_id=test,client_id=test,comm=99999999 node_disk_write_bytes={-1 if metrics['get_node_disk_write_bytes'] is None else round(metrics['get_node_disk_write_bytes'] / 30, 2)}"
        ]

        influx_client = FLInfluxDBClient(token="TokenForNCLProject", org="NCL", ip="172.17.0.1",
                                         port="8086", timeout=1_000_000)
        influx_client.connect()

        influx_client.write_new_records(bucket="running_monitoring", data=monitoring_list)

        print(monitoring_list)
        time.sleep(5)
