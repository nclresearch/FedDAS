"""
    Fl lib

    Warnings: None

    Author: Rui Sun

"""
import sys
import os
import heapq

import random
import numpy as np

import torch
from loguru import logger

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../UbiFL")))

# This progress will takes a little time and occupy 24G mem
try:
    from util_components.NAS.NAS_Bench_201.nas_201_api import NASBench201API
except ImportError:
    from UbiFL.util_components.NAS.NAS_Bench_201.nas_201_api import NASBench201API

def get_basic_model(task, model_name):
    logger.info("Initializing the model ...")
    net = None
    if task == "cv":
        if model_name == 'mnist_2nn':
            from api.utils.model_warehouse.mnist.nn.MnistNN_L2 import MnistNN_L2
            net = MnistNN_L2()
        elif model_name == 'mnist_cnn':
            from api.utils.model_warehouse.mnist.nn.MnistCNN_L2 import MnistCNN_L2
            net = MnistCNN_L2()
        elif model_name == 'lenet5':
            from api.utils.model_warehouse.mnist.lenet.LeNet5 import LeNet5
            net = LeNet5(3, 10)
        elif model_name == 'res_net':
            from api.utils.model_warehouse.cifar10.resnet.ResNet20 import ResNet20
            net = ResNet20()
        elif model_name == 'cifar10_cnn_l2':
            from api.utils.model_warehouse.cifar10.nn.CNN_L2 import CNN_L2
            net = CNN_L2()
    logger.info("Model initialized")
    return net


def get_model_from_nas(task, model_stru):
    net = None
    if task == "cv":
        from util_components.NAS.HW_NAS_Bench.hw_nas_bench_api.nas_201_models import get_cell_based_tiny_net
        net = get_cell_based_tiny_net(model_stru)  # create the network from configuration
    return net


def model_search(args):
    logger.info(
        f"Start to search network architecture under search space :{args.search_space} ... (min matrix of edgegpu_latency)")
    from util_components.NAS.HW_NAS_Bench.hw_nas_bench_api import HWNASBenchAPI as HWAPI
    config = None

    # TODO: optimize this search space name
    if args.search_space == "nasbench201":
        hw_api = HWAPI(os.path.join(args.exp_path, "util_components/NAS/HW_NAS_Bench/HW-NAS-Bench-v1_0.pickle"),
                       search_space=args.search_space)

        nas_201_api = NASBench201API(os.path.join(args.exp_path, "util_components/NAS/NAS_Bench_201/nas_201_pth/NAS-Bench-201-v1_1-096897.pth"))

        '''
            About metric_on_set list:

                if dataset == 'cifar10' and metric_on_set == 'valid':
                    dataset, metric_on_set = 'cifar10-valid', 'x-valid'
                elif dataset == 'cifar10' and metric_on_set == 'test':
                    dataset, metric_on_set = 'cifar10', 'ori-test'
                elif dataset == 'cifar10' and metric_on_set == 'train':
                   dataset, metric_on_set = 'cifar10', 'train'
                elif (dataset == 'cifar100' or dataset == 'ImageNet16-120') and metric_on_set == 'valid':
                    metric_on_set = 'x-valid'
                elif (dataset == 'cifar100' or dataset == 'ImageNet16-120') and metric_on_set == 'test':
                    metric_on_set = 'x-test'
        '''
        # Get the best model index base on highest accuracy
        best_index, highest_accuracy = nas_201_api.find_best(dataset=args.dataset, metric_on_set='test')

        # Search metrics in HW-NAS-Bench
        rank_metrics = []
        # 15625 from NAS-Bench-201 source code setting
        for idx in range(15625):
            HW_metrics = hw_api.query_by_index(idx, args.dataset)
            # Note: skip the metric who has negative value, the value close to 0 is better, this not be confirmed
            '''
                Metrics:
                    1. edgegpu_latency
                    2. edgegpu_energy
                    3. raspi4_latency
                    4. edgetpu_latency
                    5. pixel3_latency
                    6. eyeriss_latency
                    7. eyeriss_energy
                    8. eyeriss_arithmetic_intensity
                    9. fpga_latency
                    10. fpga_energy
                    11. average_hw_metric = Metric1 * Metric2 * 3 * 4..*10 
            '''
            # Filter out negative data
            if HW_metrics['edgegpu_latency'] <= 0:
                continue
            rank_metrics.append(HW_metrics['edgegpu_latency'])

        # Ranking top 10
        arr_min = heapq.nsmallest(10,rank_metrics)
        idxs_min = list(map(rank_metrics.index,arr_min))

        # TODO the result from this min(edgegpu_latencies) may not trained able, like index 15200
        selected_idx = best_index

        # if accuracy model not in the top 10 minimal inf latency, go search new optimal one
        if best_index not in idxs_min:
            logger.debug("The model with best accuracy is not in the top 10 of min inf latency, searching a new optimal model.....")
            # Search metrics in HW-NAS-Bench
            nas_201_rank_metrics = []
            # 15625 from NAS-Bench-201 source code setting
            for idx in range(15625):
                res_metrics = nas_201_api.query_meta_info_by_index(idx).get_metrics(dataset='cifar10', setname='ori-test')
                nas_201_rank_metrics.append(res_metrics['accuracy'])

            # Higher acc to lower ranking
            nas_201_rank_metrics_sorted = sorted(nas_201_rank_metrics,reverse=True)
            nas_201_rank_metrics_sorted_indexs = list(map(nas_201_rank_metrics.index, nas_201_rank_metrics_sorted))

            # Inf latency lower to higher ranking
            rank_metrics_sorted = sorted(rank_metrics)
            rank_metrics_sorted_indexs = list(map(rank_metrics.index, rank_metrics_sorted))

            idx_sum = None
            # Trade off latency and acc
            for idx in range(15625):
                if idx not in nas_201_rank_metrics_sorted_indexs or idx not in rank_metrics_sorted_indexs:
                    continue

                tmp_idx_sum = (nas_201_rank_metrics_sorted_indexs.index(idx)*args.acc_concern_rate) + (rank_metrics_sorted_indexs.index(idx)*(1-args.acc_concern_rate))

                if idx_sum is None:
                    idx_sum = tmp_idx_sum
                elif tmp_idx_sum < idx_sum:
                    idx_sum = tmp_idx_sum
                    selected_idx = idx

            logger.debug(f"nas_201_rank_metrics_sorted_indexs.index(best_idx): {nas_201_rank_metrics_sorted_indexs.index(selected_idx)}")
            logger.debug(
                f"rank_metrics_sorted_indexs.index(best_idx): {rank_metrics_sorted_indexs.index(selected_idx)}")

        hw_metrics = hw_api.query_by_index(selected_idx, args.dataset)

        logger.info(f"Selected index : {selected_idx}, HW-Metrics as fellow: ")

        logger.info("==================  Metrics from HW-NAS-Bench ==================")
        for k in hw_metrics:
            if 'average' in k:
                logger.info("{}: {}".format(k, hw_metrics[k]))
                continue
            elif "latency" in k:
                unit = "ms"
            else:
                unit = "mJ"
            logger.info("{}: {} ({})".format(k, hw_metrics[k], unit))

        logger.info("==================  Metrics from NAS-201-Bench ==================")
        nas_201_api.show(selected_idx)

        config = hw_api.get_net_config(selected_idx, args.dataset)

    elif args.search_space == "fbnet" and (args.dataset == "cifar100" or args.dataset == "ImageNet"):
        hw_api = HWAPI(os.path.join(args.exp_path, "util_components/NAS/HW_NAS_Bench/HW-NAS-Bench-v1_0.pickle"),
                       search_space=args.search_space)

        edgegpu_latencies = []
        # idxs from the HW-NAS-Bench example
        idxs = [[0] * 22, [0] * 21 + [1] * 1, [0] * 20 + [1] * 2]
        for idx in idxs:
            HW_metrics = hw_api.query_by_index(idx, args.dataset)
            edgegpu_latencies.append(HW_metrics['edgegpu_latency'])

        selected_idx = edgegpu_latencies.index(min(edgegpu_latencies))
        logger.info(f"Selected index : {selected_idx}")
        config = hw_api.get_net_config(idxs[selected_idx], args.dataset)

    logger.info(f"Network architecture searching finished, the structure is {config}")
    return config


def setup_seed(seed):
    np.random.seed(seed)
    random.seed(seed)
    # CPU
    torch.manual_seed(seed)
    # Parallel GPU
    torch.cuda.manual_seed_all(seed)
    # CPU/GPU same result
    torch.backends.cudnn.deterministic = True
    # Speed up training when the dataset is not changed alot
    # torch.backends.cudnn.benchmark = True
