"""
    Executor

    Warnings: None

    Author: Rui Sun
"""
from datetime import datetime

import requests as requests
import socket
import wandb
import sys,os
import time


sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from core.utils.general import load_yaml_conf
    from core.utils.executor_libs import *
    from api.fedavg.client import FLClient
    from core.communication.mqtt.MQTTClientManager import MQTTClientManager
except ImportError:
    from UbiFL.core.utils.general import load_yaml_conf
    from UbiFL.core.utils.executor_libs import *
    from UbiFL.api.fedavg.client import FLClient
    from UbiFL.core.communication.mqtt.MQTTClientManager import MQTTClientManager


class Executor():
    """Each executor takes certain resource to run real training.
       Each run simulates the execution of an individual mqtt_client"""
    def __init__(self, number_of_clients):

        # ======== Env Information ========
        self.dev = torch.device("cuda") if torch.cuda.is_available() and args.use_cuda else torch.device("cpu")

        self.host_id = args.worker_hostname
        self.mqtt_manager = MQTTClientManager()

        # TODO now randomly
        self.number_of_clients = number_of_clients

        # TODO this is a backup parameter
        self.this_rank = args.this_rank

    def setup_env(self,cid):
        # TODO Fix this problem

        os.chdir(args.exp_path)

        logger.info(f"(EXECUTOR:{self.this_rank}) is setting up environ ...")

        setup_seed(args.seed)

        # ============= Config Wandb =============

        wandb.init(
            project=args.wandb_project,
            name="FedAvgClient-"+self.host_id + "-r" + str(args.num_of_comm) + "-e" + str(args.local_epoch) + "-lr" + str(args.learning_rate) + "-cid_" + cid,
            config={
                "learning_rate": args.learning_rate,
                "communication": args.num_of_comm,
                "local_epoch": args.local_epoch,
                "test_batch_size": args.test_batch_size,
                "batch_size": args.batch_size
            }
        )

        load_aggregator_settings()

        # Define training in gpu or cpu
        os.environ['CUDA_VISIBLE_DEVICES'] = str(args.cuda_device)

        # set up device
        if args.use_cuda and self.dev == None:
            for i in range(torch.cuda.device_count()):
                try:
                    self.dev = torch.device('cuda:' + str(i))
                    torch.cuda.set_device(i)
                    _ = torch.rand(1).to(device=self.dev)
                    self.local_model = torch.nn.DataParallel(self.local_model)
                    logger.info(f'End up with cuda device ({self.dev})')
                    break
                except Exception as e:
                    assert i != torch.cuda.device_count() - 1, 'Can not find available GPUs'

        logger.info(f"(EXECUTOR:{self.this_rank}) finished setting up environ")

    def client_register(self):
        # Client register with server
        url = 'http://' + str(args.ps_ip) + ':' + str(args.parameter_server_client_register_port) \
              + args.parameter_server_api_client_register
        # logger.debug(url)
        params = {'hid': self.host_id}
        # headers = {'Connection': 'keep-alive', 'Accept-Encoding': 'gzip, deflate',
        #            'Accept': '*/*', 'User-Agent': 'python-requests/2.18.3'}
        # resp = requests.get(url=url, params=params, timeout=20)

        try:
            resp = requests.get(url=url, params=params, timeout=10)

            status = resp.status_code
            if (status == 200):
                cid = resp.text
                logger.info(f'Got unique mqtt_client id : {cid}')
                return cid
            # if (status == 404):
            #     status = "404 Error"
            # if (status == 403):
            #     status = "403 Error"
            # if (status == 503):
            #     status = "503 Error"

        except requests.exceptions.Timeout:
                pass

        # Connection Refused
        except requests.exceptions.ConnectionError:
            time.sleep(3)

            # Warning : watch out this recursive
            return self.client_register()



    def client_join_global_training(self,cid):
        # TODO: Client join global training
        url = 'http://' + str(args.ps_ip) + ':' + str(args.parameter_server_client_register_port) \
              + args.parameter_server_api_client_join
        # logger.debug(url)
        params = {'cid': cid}
        # headers = {'Connection': 'keep-alive', 'Accept-Encoding': 'gzip, deflate',
        #            'Accept': '*/*', 'User-Agent': 'python-requests/2.18.3'}
        resp = requests.get(url=url, params=params, timeout=20)

        res_msg = resp.text

        if res_msg != 'ok':
            logger.debug(res_msg)

    def start_a_client(self):

        cid = self.client_register()

        self.setup_env(cid)

        fl_client = FLClient(hostId=self.host_id, clientId=cid, dev=self.dev, publish_topic='/mqtt_client/publish/' + cid)
        mqtt_client = self.mqtt_manager.register(cid)
        mqtt_client.set_subscribe_topic('/aggregator/publish/' + cid)
        mqtt_client.set_publish_topic('/mqtt_client/publish/' + cid)

        mqtt_client.message_callback_add(sub=mqtt_client.subscribe_topic, callback=fl_client.on_message)
        mqtt_client.connect(args.mqtt_ip, int(args.mqtt_port))
        mqtt_client.subscribe(topic=mqtt_client.subscribe_topic,qos=2)

        # Always online
        # Warnings : join after subscribe in case the global delivered faster than join training
        self.client_join_global_training(cid)

        mqtt_client.loop_forever()

    def process_cmd(self, yaml_file):
        yaml_conf = load_yaml_conf(yaml_file)

        time_stamp = datetime.now().strftime('%Y-%m-%d_%H:%M:%S.%f')
        job_name = 'Undefined_job'
        log_path = 'running/logs'

        job_conf = {'time_stamp': time_stamp,
                    'ps_ip': yaml_conf['parameter_server']['host_ip'],
                    'mqtt_ip': yaml_conf['communication']['mqtt']['ip'],
                    'mqtt_port': yaml_conf['communication']['mqtt']['port'],
                    'parameter_server_listen_ip': yaml_conf['parameter_server']['listen_ip'],
                    'parameter_server_listen_port': yaml_conf['parameter_server']['listen_port'],
                    'parameter_server_client_register_port': yaml_conf['parameter_server']['host_port'],
                    'parameter_server_api_client_register': yaml_conf['parameter_server']['api']['client_register'],
                    'parameter_server_api_client_join': yaml_conf['parameter_server']['api']['client_join'],
                    'parameter_server_api_client_quit': yaml_conf['parameter_server']['api']['client_quit'],
                    'parameter_server_api_download_dataset': yaml_conf['parameter_server']['api']['dataset_download'],
                    'num_of_client': yaml_conf['executer']['num_of_clients'],
                    'worker_hostname': socket.gethostname(),
                    'exp_path': yaml_conf['exp_path']
                    }
        # Update args
        args.update(job_conf)

        args['log_path'] = os.path.join(log_path, job_name, time_stamp)
