"""
    Executor

    Warnings: None

    Author: Rui Sun
"""
import sys,os

import setproctitle

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from core.utils.executor_libs import *
    from api.fedavg.executor import Executor
    from core.communication.mqtt.MQTTClientManager import MQTTClientManager
except ImportError:
    from UbiFL.core.utils.executor_libs import *
    from UbiFL.api.fedavg.executor import Executor
    from UbiFL.core.communication.mqtt.MQTTClientManager import MQTTClientManager

# for multi executor
def run(num_clients_to_run, rank=0):
    setproctitle.setproctitle("UbiFL_Executor "+str(rank))
    exec = Executor(num_clients_to_run)
    exec.start_a_client()

if __name__=="__main__":
    # number_of_clients = int(max((args.num_of_clients * args.client_fraction) / args.num_of_executors, 1))

    '''
    address problem: Cannot re-initialize CUDA in forked subprocess. To use CUDA with multiprocessing
    https://blog.csdn.net/slamdunkofkd/article/details/119670914
    '''
    # torch.multiprocessing.set_start_method('spawn')

    logger.info(f"Creating client...")

    # Leave for multi mqtt_client setting
    # num_clients_to_run = args.num_of_clients
    # for multi executor
    # for i in range(num_clients_to_run):
    #     logger.info(f"Creating the {i + 1} mqtt_client...")
    #     p = torch.multiprocessing.Process(target=run, args=(num_clients_to_run,i))
    #     p.start()

    num_clients_to_run = 1
    run(num_clients_to_run)


