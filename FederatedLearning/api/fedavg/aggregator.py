"""
    Aggregator

    Warnings: None

    Author: Rui Sun
"""
import pickle
import threading
import time

import wandb
import torch.backends.cudnn
from tqdm import tqdm
from torch import nn

import sys, os

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from core.utils.general import test_mkdir
    from core.utils.arguments import args
    from api.fedavg.client_manager import ClientManager
    from api.data_preprocessing.data_loader import get_data_loader
    from core.communication.mqtt.MQTTClientManager import MQTTClientManager
    from core.utils.aggregator_libs import *
    from util_components.monitor.prometheus.prome_query_4_cloud import PrometheusQuery4Cloud
    from DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient
except ImportError:
    from UbiFL.core.utils.general import test_mkdir
    from UbiFL.core.utils.arguments import args
    from UbiFL.api.fedavg.client_manager import ClientManager
    from UbiFL.api.data_preprocessing.data_loader import get_data_loader
    from UbiFL.core.communication.mqtt.MQTTClientManager import MQTTClientManager
    from UbiFL.core.utils.aggregator_libs import *
    from UbiFL.util_components.monitor.prometheus.prome_query_4_cloud import PrometheusQuery4Cloud
    from UbiFL.DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient


class Aggregator:
    """This centralized aggregator collects training/testing feedbacks from executors"""

    def __init__(self):
        # ======== Env Information ========
        self.datasets = None
        self.client_manager = self.init_client_manager()
        self.tqdm_bar = None
        self.dev = torch.device("cuda") if torch.cuda.is_available() and args.use_cuda else torch.device("cpu")
        # TODO put it in args, now select all
        # self.num_of_sample = self.client_manager.selected_num_of_clients

        # ======== Model and Data ========
        # init model
        if args.model_from == "basic":
            self.global_model = get_basic_model(task=args.task, model_name=args.model_name)
        elif args.model_from == "nas":
            self.model_structure = model_search(args=args)
            self.global_model = get_model_from_nas(task=args.task, model_stru=self.model_structure)

        self.global_parameters = {}
        for key, var in self.global_model.state_dict().items():
            self.global_parameters[key] = var.clone()

        self.sum_client_model_parameters = None

        self.locker = threading.Lock()
        self.is_iid = False

        # ======== Communication ========
        self.mqtt_manager = MQTTClientManager()

        # ======== Runtime Information ========
        self.total_communication_round = args.num_of_comm
        self.current_communication_round = 0
        self.loss_func = nn.CrossEntropyLoss()
        self.is_under_training = False

        # ======== result ========
        self.all_results = []
        self.all_accuracy = []
        self.all_loss = []

        # ======== Monitoring ========
        # InfluxDB client
        if args.is_monitoring:
            self.influx_client = FLInfluxDBClient(token=args.influxdb_token, org=args.influxdb_org, ip=args.influxdb_ip,
                                                  port=args.influxdb_port, timeout=1_000_000)
            self.influx_client.connect()
            # self.monitoring_locker = threading.Lock()
            self.agg_monitor = AggregatorMonitor()
            self.prome_monitor_thread = threading.Thread(target=self.monitoring)

            # self.prome_query = PrometheusQuery4Cloud(args.prometheus_ip, args.prometheus_port)
            self.prome_monitor_thread.start()

    def setup_env(self):
        # TODO Fix this problem
        os.chdir(args.exp_path)

        logger.info("(AGGREGATOR) is setting up environ ...")
        test_mkdir(args.checkpoints_save_path)
        load_aggregator_settings()

        setup_seed(args.seed)
        # ============= Config Wandb =============

        wandb.init(
            project=args.wandb_project,
            name="FedAVG-Aggregator-r" + str(args.num_of_comm) + "-e" + str(args.local_epoch) + "-lr" + str(
                args.learning_rate) + "-model" + args.model_name,
            config={
                "learning_rate": args.learning_rate,
                "communication": args.num_of_comm,
                "local_epoch": args.local_epoch,
                "test_batch_size": args.test_batch_size,
                "batch_size": args.batch_size
            }
        )

        # Define training in gpu or cpu
        os.environ['CUDA_VISIBLE_DEVICES'] = str(args.cuda_device)

        # set up device
        if args.use_cuda and self.dev == None:
            for i in range(torch.cuda.device_count()):
                try:
                    self.dev = torch.device('cuda:' + str(i))
                    torch.cuda.set_device(i)
                    # _ = torch.rand(1).to(device=self.dev)
                    self.global_model = torch.nn.DataParallel(self.global_model)
                    logger.info(f'End up with cuda device ({self.dev})')
                    break
                except Exception as e:
                    assert i != torch.cuda.device_count() - 1, 'Can not find available GPUs'

        self.global_model = self.global_model.to(self.dev)

        logger.info("(AGGREGATOR) finished setting up environ")

    def broadcast_models(self):
        """Push the latest model to executors"""
        # clean result queue means the an new turn is start
        self.all_results.clear()
        self.client_manager.clear_finished_clients_queue()

        # clean parameter saver
        self.sum_client_model_parameters = None

        payload_to_clients = {
            "communication": {
                "current_communication_round": self.current_communication_round,
                "total_communication_round": self.total_communication_round
            },
            "global_model": {
                "params": self.global_parameters
            },
            "sys": {
                "dev" : "cuda" if torch.cuda.is_available() and args.use_cuda else "cpu"
            }
        }

        if args.is_monitoring:
            payload_to_clients["monitoring"] = {
                "sent_time": str(int(round(time.time() * 1000)))
            }

        # if network from nas, should let client create net base on this structure
        if args.model_from == "nas":
            payload_to_clients["global_model"]["model_structure"] = self.model_structure

        # init tqdm
        self.tqdm_bar = tqdm(total=len(self.mqtt_manager.all_mqtt_clients), desc="Client local training ...")

        # Loop online all clients
        # TODO Randomly choose part of mqtt_client
        for cid, mq_client in self.mqtt_manager.all_mqtt_clients.items():
            mq_client.message_callback_add(sub=mq_client.subscribe_topic, callback=self.on_message)
            mq_client.connect(args.mqtt_ip, int(args.mqtt_port))
            mq_client.subscribe(topic=mq_client.subscribe_topic, qos=2)
            mq_client.publish(topic=mq_client.publish_topic, qos=2, payload=payload_to_clients)
            mq_client.loop_start()

    def aggregate_all_local_models(self):

        logger.info("(AGGREGATOR) is aggregating all sub-models ...")

        for key in self.global_parameters.keys():
            # self.global_parameters[key] = torch.div(self.sum_client_model_parameters[key], self.num_of_clients)
            self.global_parameters[key] = (
                        self.sum_client_model_parameters[key] / self.client_manager.selected_num_of_clients)

        self.global_model.load_state_dict(self.global_parameters)

        logger.info("(AGGREGATOR) finished aggregate all sub-models")

    def init_client_manager(self):
        """
            1. Random mqtt_client sampler
                - it selects participants randomly in each round
                - [Ref]: https://arxiv.org/abs/1902.01046
        """

        # sample_mode: random or kuiper
        client_manager = ClientManager(args.client_sample_mode, args=args)

        return client_manager

    def on_message(self, client, userdata, msg):
        # Callback func

        # Record working progress
        if self.tqdm_bar is not None:
            self.tqdm_bar.update()

        client_payload = pickle.loads(msg.payload)
        logger.debug(f'Got result ! from {client_payload["clientId"]}')

        if args.is_monitoring:
            # ============= Communication delay monitoring (Network) =============
            delta_t = int(round(time.time() * 1000)) - int(client_payload["monitoring"]["sent_time"])
            data = f"server_running_status,device_id={client_payload['clientId']},comm={self.current_communication_round} comm_delay={delta_t}"

            self.influx_client.write_new_records(bucket="running_monitoring", data=data)

        local_parameters = client_payload["parameters"]["weight_parameters"]["local"]

        # Saving result to queue
        self.locker.acquire()
        self.all_results.append(client_payload)
        self.client_manager.finished_training_clients.append(client_payload["clientId"])

        # Save all parameters
        # Warnings: may waste memory

        if self.sum_client_model_parameters is None:
            self.sum_client_model_parameters = {}
            for key, var in local_parameters.items():
                self.sum_client_model_parameters[key] = var.clone()
        else:
            for key, param in local_parameters.items():
                self.sum_client_model_parameters[key] += param

        # logger.debug(f"local_parameters.keys() {local_parameters.keys()}")
        # logger.debug(f"self.sum_client_model_parameters.keys() {self.sum_client_model_parameters.keys()}")

        self.locker.release()

        # if self.total_communication_round == self.current_communication_round:
        #     self.is_under_training = False

    def __test(self, model):

        model.load_state_dict(self.global_parameters, strict=True)
        model.eval()
        total_correct = 0.
        total_test_loss = 0.
        total_size = 0
        counter = 0
        with torch.no_grad():
            for data, label in self.datasets["test_data_global"]:
                counter += 1
                data, label = data.to(self.dev), label.to(self.dev)
                preds = model(data)
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
            return total_correct * 100., total_test_loss

    def run(self):

        # TODO: run sub clients and register them
        self.setup_env()

        for i in tqdm(range(args.num_of_comm), desc=f"Communication round ... :"):
            # logger.debug(f"self.current_communication_round {self.current_communication_round}")
            # logger.debug(f"self.client_finished_training.qsize() {len(self.client_finished_training)}")

            while True:
                if self.current_communication_round == 0 and len(
                        self.client_manager.feasible_clients) == self.client_manager.selected_num_of_clients:
                    # Data partition for first round
                    # Patatition data for every mqtt_client and save them in file
                    # TODO should support mqtt_client use their local data

                    # Warnings: May challenge memory
                    self.datasets = get_data_loader(dataset_name=args.dataset,
                                                    cids=self.client_manager.feasible_clients,
                                                    args=args)
                    break
                # TODO now only support all mqtt_client should return their results
                elif len(self.client_manager.finished_training_clients) == self.client_manager.selected_num_of_clients:
                    logger.debug('All workers finished')
                    self.aggregate_all_local_models()
                    break

                time.sleep(1)

            if self.current_communication_round == 1:
                self.is_under_training = True

            self.current_communication_round += 1
            # logger.info(f"Communicate round {self.current_communication_round}")

            # Deliver model
            self.broadcast_models()

            with torch.no_grad():
                if (i + 1) % args.test_freq == 0:
                    acc, loss = self.__test(self.global_model)
                    loss = round(loss, 4)
                    logger.info(f'Aggregator 0 global model accuracy: {acc}')
                    logger.info(f'Aggregator 0 global model loss: {loss}')
                    self.all_accuracy.append(acc)
                    self.all_loss.append(loss)

                    wandb.log({
                        "global_model_test/Acc": acc,
                        "global_model_test/Loss": loss
                    })

                    # >>>>>>>>>> Monitoring >>>>>>>>>>>
                    if args.is_monitoring:
                        sequence = [
                            f"global_model_result,device_id=aggregator test_acc={acc}",
                            f"global_model_result,device_id=aggregator test_loss={loss}"
                        ]
                        self.influx_client.write_new_records(bucket="running_monitoring", data=sequence)
                    # <<<<<<<<<< Monitoring <<<<<<<<<<

            if (i + 1) % args.save_freq == 0:
                torch.save(self.global_model, os.path.join(args.save_path,
                                                           f'{args.model_name}_num_comm{i}_E{args.local_epoch}_B{args.batch_size}'
                                                           f'_lr{args.learning_rate}_num_clients{args.num_of_clients}_cf{args.client_fraction}'))

        torch.save(self.all_accuracy, 'cache/accuracy/aggregator.pkl')
        torch.save(self.all_loss, 'cache/loss/aggregator.pkl')

        self.is_under_training = False
        self.all_accuracy.clear()
        self.all_loss.clear()
        self.client_manager.clear_finished_clients_queue()
        self.agg_monitor.close()


        if args.is_monitoring:
            self.influx_client.close()

    '''
        Monitoring
    '''

    def monitoring(self):
        while True:
            if self.is_under_training:
                sequence = self.agg_monitor.prometheus_monitor(self.current_communication_round)
                # logger.info(sequence)
                self.agg_monitor.record_metrics(sequence)
            time.sleep(5)

'''
    Monitoring
'''
class AggregatorMonitor(object):
    def __init__(self,hostId="aggregator"):
        self.hostId = hostId

        self.influx_client = FLInfluxDBClient(token=args.influxdb_token, org=args.influxdb_org, ip=args.influxdb_ip,
                                                  port=args.influxdb_port, timeout=1_000_000)
        self.influx_client.connect()
        self.prome_query = PrometheusQuery4Cloud(args.prometheus_ip, args.prometheus_port)


    def prometheus_monitor(self, current_communication_round):
        if self.prome_query.pod_name is None:
            self.prome_query.update_pod_name()

            # self.monitoring_locker.acquire()
        metrics = self.get_metrics()

        monitoring_list = [
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_cpu_average={-1 if metrics['get_node_cpu_usage'] is None else round(metrics['get_node_cpu_usage'], 4)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} pod_cpu_usage={-1 if metrics['get_node_cpu_usage'] is None else round(metrics['get_pod_cpu_usage'], 4)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_mem_usage_average={-1 if metrics['get_node_mem_usage'] is None else round(metrics['get_node_mem_usage'], 4)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} pod_mem_usage_average={-1 if metrics['get_pod_mem_usage'] is None else round(metrics['get_pod_mem_usage'], 4)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} pod_disk_read_bytes={-1 if metrics['get_pod_disk_read_bytes'] is None else round(metrics['get_pod_disk_read_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} pod_disk_write_bytes={-1 if metrics['get_pod_disk_write_bytes'] is None else round(metrics['get_pod_disk_write_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} pod_network_receive_bytes={-1 if metrics['get_pod_network_receive_bytes'] is None else round(metrics['get_pod_network_receive_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} pod_network_transmit_bytes={-1 if metrics['get_pod_network_transmit_bytes'] is None else round(metrics['get_pod_network_transmit_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_network_receive_packets={-1 if metrics['get_node_network_receive_packets'] is None else metrics['get_node_network_receive_packets']}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_network_receive_errors={-1 if metrics['get_node_network_receive_errors'] is None else metrics['get_node_network_receive_errors']}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_network_transmit_errors={-1 if metrics['get_node_network_transmit_errors'] is None else metrics['get_node_network_transmit_errors']}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_network_receive_bytes={-1 if metrics['get_node_network_receive_bytes'] is None else round(metrics['get_node_network_receive_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_network_transmit_bytes={-1 if metrics['get_node_network_transmit_bytes'] is None else round(metrics['get_node_network_transmit_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_disk_read_bytes={-1 if metrics['get_node_disk_read_bytes'] is None else round(metrics['get_node_disk_read_bytes'] / 30, 2)}",
            f"server_running_status,device_id={self.hostId},comm={current_communication_round} node_disk_write_bytes={-1 if metrics['get_node_disk_write_bytes'] is None else round(metrics['get_node_disk_write_bytes'] / 30, 2)}"]
        # self.monitoring_locker.release()
        # logger.info(monitoring_list)
        return monitoring_list

    def get_metrics(self):
        return {
            "get_node_cpu_usage": self.prome_query.get_node_cpu_usage('30s'),
            "get_pod_cpu_usage": self.prome_query.get_pod_cpu_usage('30s'),
            "get_node_mem_usage": self.prome_query.get_node_mem_usage,
            "get_pod_mem_usage": self.prome_query.get_pod_mem_usage,
            "get_pod_disk_read_bytes": self.prome_query.get_pod_disk_read_bytes('30s'),
            "get_pod_disk_write_bytes": self.prome_query.get_pod_disk_write_bytes('30s'),
            "get_pod_network_receive_bytes": self.prome_query.get_pod_network_receive_bytes('30s'),
            "get_pod_network_transmit_bytes": self.prome_query.get_pod_network_transmit_bytes('30s'),
            "get_node_network_receive_packets": self.prome_query.get_node_network_receive_packets('30s'),
            "get_node_network_receive_errors": self.prome_query.get_node_network_receive_errors('30s'),
            "get_node_network_transmit_errors": self.prome_query.get_node_network_transmit_errors('30s'),
            "get_node_network_receive_bytes": self.prome_query.get_node_network_receive_bytes('30s'),
            "get_node_network_transmit_bytes": self.prome_query.get_node_network_transmit_bytes('30s'),
            "get_node_disk_read_bytes": self.prome_query.get_node_disk_read_bytes('30s'),
            "get_node_disk_write_bytes": self.prome_query.get_node_disk_write_bytes('30s')
        }

    def record_metrics(self,sequence):
        self.influx_client.write_new_records(bucket="running_monitoring", data=sequence)

    def close(self):
        self.influx_client.close()