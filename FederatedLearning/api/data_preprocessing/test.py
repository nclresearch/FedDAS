# import argparse
# import logging
# import os
# import random
# import socket
# import sys
#
# import numpy as np
# import torch
#
# sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../")))
#
# from fedml.data.cifar10.data_loader import load_partition_data_cifar10
#
#
#
# def add_args(parser):
#     """
#     parser : argparse.ArgumentParser
#     return a parser added with args required by fit
#     """
#     # Training settings
#     parser.add_argument('--dataset', type=str, default="cifar10")
#
#     # default arguments
#     parser.add_argument("--data_dir", type=str, default='../../dataset/data/cifar10')
#
#     # default arguments
#     parser.add_argument("--partition_method", type=str, default='homo')
#
#     parser.add_argument("--partition_alpha", type=float, default=1)
#
#     parser.add_argument("--client_num_in_total", type=int, default=2)
#
#     parser.add_argument("--batch_size", type=int, default=64)
#
#     args = parser.parse_args()
#     return args
#
#
# def load_data(args, dataset_name):
#     if dataset_name == "cifar10":
#         data_loader = load_partition_data_cifar10
#
#
#         train_data_num, test_data_num, train_data_global, test_data_global, \
#         train_data_local_num_dict, train_data_local_dict, test_data_local_dict, \
#         class_num = data_loader(args.dataset, args.data_dir, args.partition_method,
#                                 args.partition_alpha, args.client_num_in_total, args.batch_size)
#     dataset = [train_data_num, test_data_num, train_data_global, test_data_global,
#                train_data_local_num_dict, train_data_local_dict, test_data_local_dict, class_num]
#     return dataset
#
# if __name__ == "__main__":
#
#     # from fedml.data.cifar10.data_loader import load_cifar10_data
#     #
#     # x1,x2,y1,y2 = load_cifar10_data('../../dataset/data/cifar10')
#
#
#     parser = argparse.ArgumentParser(description="UbiFL")
#
#     args = add_args(parser)
#
#     load_data(args,args.dataset)
#
#     # Set the random seed. The np.random seed determines the dataset partition.
#     # The torch_manual_seed determines the initial weight.
#     # We fix these two, so that we can reproduce the result.
#     random.seed(0)
#     np.random.seed(0)
#     torch.manual_seed(0)
#     torch.cuda.manual_seed_all(0)
#
#
#     # load data
#     dataset = load_data(args, args.dataset)
#     [train_data_num, test_data_num, train_data_global, test_data_global,
#      train_data_local_num_dict, train_data_local_dict, test_data_local_dict, class_num] = dataset
#
#     print(train_data_local_num_dict)
#
#     # for k,v in test_data_local_dict.items():
#     #     print(k)
#     #     for data,label in v:
#     #         print(label)
#     #         break
#     #
#     #     for d,v in test_data_global:
#     #         print(v)
#     #         break
#
#
#     # for x,y in train_data_local_dict[0].dataset:
#     #     print(y)
#     # print(train_data_local_dict[0].dataset)
#
#     # np.save('./testdata.npy', train_data_local_dict[0].dataset)





# def http():
#     url = 'http://localhost:8088/client/register'
#     # logger.debug(url)
#     params = {'hid': "test"}
#     # headers = {'Connection': 'keep-alive', 'Accept-Encoding': 'gzip, deflate',
#     #            'Accept': '*/*', 'User-Agent': 'python-requests/2.18.3'}
#     # resp = requests.get(url=url, params=params)
#
#     try:
#         resp = requests.get(url=url, params=params, timeout=10)
#
#         status = resp.status_code
#         if (status == 200):
#             status = "Connection Successful"
#         if (status == 404):
#             status = "404 Error"
#         if (status == 403):
#             status = "403 Error"
#         if (status == 503):
#             status = "503 Error"
#         print(status)
#
#     except requests.exceptions.Timeout:
#         status = "Connection Timed Out"
#         print(status)
#
#
#     except requests.exceptions.ConnectionError:
#         status = "Connection Refused"
#         print(status)
#         time.sleep(3)
#         return http()
#
#
# import requests
# import time
# if __name__ == '__main__':
#     expr = [
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time)  |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage")  |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_memory_usage: t2}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t3 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_disk_read_bytes: t3}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t4 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_disk_write_bytes: t4}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t5 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_network_receive_bytes: t5}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t6 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_network_transmit_bytes: t6}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t7 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_network_receive_bytes: t7}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t8 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_network_transmit_bytes: t8}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t9 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_cpu_user_usage_avg") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_cpu_user_usage_avg: t9}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t10 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_cpu_sys_usage_avg") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_cpu_sys_usage_avg: t10}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t11 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_memory_usage: t11}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t12 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_disk_read_bytes: t12}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_disk_write_bytes: t13}, on: ["_time"]) |> yield (name: "last")',
#
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu1") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu1: t14}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu2") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu2: t15}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu3") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu3: t16}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu4") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu4: t17}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_gpu") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_gpu: t18}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_ram") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_ram: t19}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_temperature_cpu") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_temperature_cpu: t20}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_temperature_gpu") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_temperature_gpu: t21}, on: ["_time"]) |> yield (name: "last")',
#         'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_power_cur") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_power_cur: t22}, on: ["_time"]) |> yield (name: "last")',
#
#     ]
import sys, os

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))

from DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient

if __name__ == '__main__':
    # a = [5,2,7,3,100,2,24,51,5632,6,23,54]
    # b = sorted(a,reverse=True)
    # c = list(map(a.index,b))
    # # c = list(map(b.index, a))
    # print(b)
    # print(c)
    #
    # import heapq
    #
    # d = heapq.nsmallest(3, a)
    # e = list(map(a.index, d))
    # print(e)
    # for a in range(3):
    #     print(a)

    influx_client = FLInfluxDBClient(token="TokenForNCLProject", timeout=1000000000000, org='NCL', ip="10.70.31.211",
                                     port="12186")

    expr =  'start_time=-10s t1 = from(bucket: "running_monitoring") |> range(start: start_time)  |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage")  |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_memory_usage: t2}, on: ["_time"]) |> yield (name: "last")'

    large_stream = influx_client.query_stream(expr)
    for record in large_stream:
        print(record)
    large_stream.close()