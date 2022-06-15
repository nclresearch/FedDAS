import sys, os

sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))

from DAO.influxdb.FLInfluxDBClient import FLInfluxDBClient
from loguru import logger
import pandas as pd
from tqdm import tqdm

iot_port = ['12186', '12286', '12386', '12486', '12586']

# iot_port =['12186','12286']

dataset_df = pd.DataFrame()
res = None

# start_time = '2022-06-01T12:40:00Z'
# stop_time = '2022-06-02T20:39:00Z'

start_time = '2022-06-04T00:16:00Z'
stop_time = '2022-06-04T22:27:00Z'

counter = 0
for port in tqdm(iot_port):
    counter += 1
    influx_client = FLInfluxDBClient(token="TokenForNCLProject", timeout=1000000000000, org='NCL', ip="10.70.31.211",
                                     port=port)

    expr = [
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time)  |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage")  |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_memory_usage: t2}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t3 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_disk_read_bytes: t3}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t4 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_disk_write_bytes: t4}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t5 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_network_receive_bytes: t5}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t6 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_network_transmit_bytes: t6}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t7 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_network_receive_bytes: t7}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t8 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_network_transmit_bytes: t8}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t9 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_cpu_user_usage_avg") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_cpu_user_usage_avg: t9}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t10 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_cpu_sys_usage_avg") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_cpu_sys_usage_avg: t10}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t11 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_memory_usage: t11}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t12 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_disk_read_bytes: t12}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_disk_write_bytes: t13}, on: ["_time"]) |> yield (name: "last")',

        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t14 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu1") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu1: t14}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t15 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu2") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu2: t15}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t16 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu3") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu3: t16}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t17 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jetson_stats_cpu4") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jetson_stats_cpu4: t17}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t18 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_gpu") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_gpu: t18}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t19 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_ram") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_ram: t19}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t20 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_temperature_cpu") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_temperature_cpu: t20}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t21 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_temperature_gpu") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_temperature_gpu: t21}, on: ["_time"]) |> yield (name: "last")',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t22 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "jtop_stats_power_cur") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, jtop_stats_power_cur: t22}, on: ["_time"]) |> yield (name: "last")',

    ]

    # query comm delay
    expr_comm_delay = 'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field  == "comm_delay") |> sort(columns:["_time"])  join(tables: {container_cpu_average: t1, comm_delay: t2}, on: ["comm"])'

    # query epoch delay
    expr_epoch_delay_avg = 'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field  == "epoch_delay_avg") |> sort(columns:["_time"])  join(tables: {container_cpu_average: t1, epoch_delay_avg: t2}, on: ["comm"])'

    print("start to query t1_t2")
    t1_t2 = influx_client.query(expr[0])
    print("start to query t1_t3")
    t1_t3 = influx_client.query(expr[1])
    print("start to query t1_t4")
    t1_t4 = influx_client.query(expr[2])
    print("start to query t1_t5")
    t1_t5 = influx_client.query(expr[3])
    print("start to query t1_t6")
    t1_t6 = influx_client.query(expr[4])
    print("start to query t1_t7")
    t1_t7 = influx_client.query(expr[5])
    print("start to query t1_t8")
    t1_t8 = influx_client.query(expr[6])
    print("start to query t1_t9")
    t1_t9 = influx_client.query(expr[7])
    print("start to query t1_t10")
    t1_t10 = influx_client.query(expr[8])
    print("start to query t1_t11")
    t1_t11 = influx_client.query(expr[9])
    print("start to query t1_t12")
    t1_t12 = influx_client.query(expr[10])
    print("start to query t1_t13")
    t1_t13 = influx_client.query(expr[11])

    print("start to query t1_t14")
    t1_t14 = influx_client.query(expr[12])
    print("start to query t1_t15")
    t1_t15 = influx_client.query(expr[13])
    print("start to query t1_t16")
    t1_t16 = influx_client.query(expr[14])
    print("start to query t1_t17")
    t1_t17 = influx_client.query(expr[15])
    print("start to query t1_t18")
    t1_t18 = influx_client.query(expr[16])
    print("start to query t1_t19")
    t1_t19 = influx_client.query(expr[17])
    print("start to query t1_t20")
    t1_t20 = influx_client.query(expr[18])
    print("start to query t1_t21")
    t1_t21 = influx_client.query(expr[19])
    print("start to query t1_t22")
    t1_t22 = influx_client.query(expr[20])

    # combine to the main df
    comm_delay_df = influx_client.query(expr_comm_delay)
    deplay_delay_avg_df = influx_client.query(expr_epoch_delay_avg)

    if counter == 1:
        print(t1_t2)
        dataset_df = t1_t2[
            ['_value_container_cpu_average', '_value_container_memory_usage', 'comm_container_cpu_average',
             'device_id_container_cpu_average']].copy()

        dataset_df['_value_container_disk_read_bytes'] = t1_t3[['_value_container_disk_read_bytes']].copy()
        dataset_df['_value_container_disk_write_bytes'] = t1_t4[['_value_container_disk_write_bytes']].copy()
        dataset_df['_value_container_network_receive_bytes'] = t1_t5[['_value_container_network_receive_bytes']].copy()
        dataset_df['_value_container_network_transmit_bytes'] = t1_t6[
            ['_value_container_network_transmit_bytes']].copy()
        dataset_df['_value_node_network_receive_bytes'] = t1_t7[['_value_node_network_receive_bytes']].copy()
        dataset_df['_value_node_network_transmit_bytes'] = t1_t8[['_value_node_network_transmit_bytes']].copy()
        dataset_df['_value_node_cpu_user_usage_avg'] = t1_t9[['_value_node_cpu_user_usage_avg']].copy()
        dataset_df['_value_node_cpu_sys_usage_avg'] = t1_t10[['_value_node_cpu_sys_usage_avg']].copy()
        dataset_df['_value_node_memory_usage'] = t1_t11[['_value_node_memory_usage']].copy()
        dataset_df['_value_node_disk_read_bytes'] = t1_t12[['_value_node_disk_read_bytes']].copy()
        dataset_df['_value_node_disk_write_bytes'] = t1_t13[['_value_node_disk_write_bytes']].copy()

        dataset_df['_value_jetson_stats_cpu1'] = t1_t14[['_value_jetson_stats_cpu1']].copy()
        dataset_df['_value_jetson_stats_cpu2'] = t1_t15[['_value_jetson_stats_cpu2']].copy()
        dataset_df['_value_jetson_stats_cpu3'] = t1_t16[['_value_jetson_stats_cpu3']].copy()
        dataset_df['_value_jetson_stats_cpu4'] = t1_t17[['_value_jetson_stats_cpu4']].copy()
        dataset_df['_value_jtop_stats_gpu'] = t1_t18[['_value_jtop_stats_gpu']].copy()
        dataset_df['_value_jtop_stats_ram'] = t1_t19[['_value_jtop_stats_ram']].copy()
        dataset_df['_value_jtop_stats_temperature_cpu'] = t1_t20[['_value_jtop_stats_temperature_cpu']].copy()
        dataset_df['_value_jtop_stats_temperature_gpu'] = t1_t21[['_value_jtop_stats_temperature_gpu']].copy()
        dataset_df['_value_jtop_stats_power_cur'] = t1_t22[['_value_jtop_stats_power_cur']].copy()

        dataset_df['_value_comm_delay'] = comm_delay_df[['_value_comm_delay']].copy()
        dataset_df['_value_epoch_delay_avg'] = deplay_delay_avg_df[['_value_epoch_delay_avg']].copy()
        print(dataset_df.shape)
    else:
        df = t1_t2[['_value_container_cpu_average', '_value_container_memory_usage', 'comm_container_cpu_average',
                    'device_id_container_cpu_average']].copy()
        df['_value_container_disk_write_bytes'] = t1_t4[['_value_container_disk_write_bytes']].copy()
        df['_value_container_network_receive_bytes'] = t1_t5[['_value_container_network_receive_bytes']].copy()
        df['_value_container_network_transmit_bytes'] = t1_t6[['_value_container_network_transmit_bytes']].copy()
        df['_value_node_network_receive_bytes'] = t1_t7[['_value_node_network_receive_bytes']].copy()
        df['_value_node_network_transmit_bytes'] = t1_t8[['_value_node_network_transmit_bytes']].copy()
        df['_value_node_cpu_user_usage_avg'] = t1_t9[['_value_node_cpu_user_usage_avg']].copy()
        df['_value_node_cpu_sys_usage_avg'] = t1_t10[['_value_node_cpu_sys_usage_avg']].copy()
        df['_value_node_memory_usage'] = t1_t11[['_value_node_memory_usage']].copy()
        df['_value_node_disk_read_bytes'] = t1_t12[['_value_node_disk_read_bytes']].copy()
        df['_value_node_disk_write_bytes'] = t1_t13[['_value_node_disk_write_bytes']].copy()

        df['_value_jetson_stats_cpu1'] = t1_t14[['_value_jetson_stats_cpu1']].copy()
        df['_value_jetson_stats_cpu2'] = t1_t15[['_value_jetson_stats_cpu2']].copy()
        df['_value_jetson_stats_cpu3'] = t1_t16[['_value_jetson_stats_cpu3']].copy()
        df['_value_jetson_stats_cpu4'] = t1_t17[['_value_jetson_stats_cpu4']].copy()
        df['_value_jtop_stats_gpu'] = t1_t18[['_value_jtop_stats_gpu']].copy()
        df['_value_jtop_stats_ram'] = t1_t19[['_value_jtop_stats_ram']].copy()
        df['_value_jtop_stats_temperature_cpu'] = t1_t20[['_value_jtop_stats_temperature_cpu']].copy()
        df['_value_jtop_stats_temperature_gpu'] = t1_t21[['_value_jtop_stats_temperature_gpu']].copy()
        df['_value_jtop_stats_power_cur'] = t1_t22[['_value_jtop_stats_power_cur']].copy()

        df['_value_comm_delay'] = comm_delay_df[['_value_comm_delay']].copy()
        df['_value_epoch_delay_avg'] = deplay_delay_avg_df[['_value_epoch_delay_avg']].copy()

        dataset_df = pd.concat([dataset_df, df], axis=0, ignore_index=True)
        print(dataset_df.shape)
    dataset_df.to_csv("edge_new_dataset-"+port+".csv")


print(dataset_df.shape)
dataset_df.to_csv("edge_new_dataset.csv")

# print(comm_delay_df.shape)
# comm_delay_df.to_csv("edge_comm_delay_dataset.csv")
#
# print(comm_delay_df.shape)
# deplay_delay_avg_df.to_csv("edge_epoch_delay_avg_dataset.csv")