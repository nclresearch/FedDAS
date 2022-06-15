from monitoring.anomaly_detector import FLInfluxDBClient

# iot_port =['12186','12286','12386','12486','12586']
iot_port = ['12286', '12486', '12586']

dataset_df = pd.DataFrame()

start_time = '2022-01-22T21:43:36Z'
stop_time = '2022-01-23T11:45:00Z'

counter = 0
for port in tqdm(iot_port):
    counter += 1
    influx_client = FLInfluxDBClient(token="TokenForNCLWWW2022DemoPaperProject", org='NCL', ip="10.70.31.176",
                                     port=port)

    expr = [
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time)  |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage")  |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_memory_usage: t2}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t3 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_disk_read_bytes: t3}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t4 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_disk_write_bytes: t4}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t5 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_network_receive_bytes: t5}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t6 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, container_network_transmit_bytes: t6}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t7 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_network_receive_bytes: t7}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t8 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_network_transmit_bytes: t8}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t9 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_cpu_user_usage_avg") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_cpu_user_usage_avg: t9}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t10 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_cpu_sys_usage_avg") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_cpu_sys_usage_avg: t10}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t11 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_memory_usage") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_memory_usage: t11}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t12 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_disk_read_bytes: t12}, on: ["_time"])',
        'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t13 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "node_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {container_cpu_average: t1, node_disk_write_bytes: t13}, on: ["_time"])'
    ]

    # query comm delay
    expr_comm_delay = 'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field  == "comm_delay") |> sort(columns:["_time"])  join(tables: {container_cpu_average: t1, comm_delay: t2}, on: ["comm"])'

    # query epoch delay
    expr_epoch_delay_avg = 'start_time=' + start_time + ' stop_time=' + stop_time + ' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field == "container_cpu_usage") |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time) |> filter(fn: (r) => r._measurement == "client_running_status" and r._field  == "epoch_delay_avg") |> sort(columns:["_time"])  join(tables: {container_cpu_average: t1, epoch_delay_avg: t2}, on: ["comm"])'

    t1_t2 = influx_client.query(expr[0])
    t1_t3 = influx_client.query(expr[1])
    t1_t4 = influx_client.query(expr[2])
    t1_t5 = influx_client.query(expr[3])
    t1_t6 = influx_client.query(expr[4])
    t1_t7 = influx_client.query(expr[5])
    t1_t8 = influx_client.query(expr[6])
    t1_t9 = influx_client.query(expr[7])
    t1_t10 = influx_client.query(expr[8])
    t1_t11 = influx_client.query(expr[9])
    t1_t12 = influx_client.query(expr[10])
    t1_t13 = influx_client.query(expr[11])

    # combine to the main df
    comm_delay_df = influx_client.query(expr_comm_delay)
    deplay_delay_avg_df = influx_client.query(expr_epoch_delay_avg)

    if counter == 1:

        dataset_df = t1_t2[
            ['_value_container_cpu_average', '_value_container_memory_usage', 'comm_container_cpu_average',
             'device_id_container_cpu_average']].copy()
        dataset_df['_value_node_mem_usage_average'] = t1_t3[['_value_container_disk_read_bytes']].copy()
        dataset_df['_value_pod_mem_usage_average'] = t1_t4[['_value_container_disk_write_bytes']].copy()
        dataset_df['_value_pod_disk_read_bytes'] = t1_t5[['_value_container_network_receive_bytes']].copy()
        dataset_df['_value_pod_disk_write_bytes'] = t1_t6[['_value_container_network_transmit_bytes']].copy()
        dataset_df['_value_pod_network_receive_bytes'] = t1_t7[['_value_node_network_receive_bytes']].copy()
        dataset_df['_value_pod_network_transmit_bytes'] = t1_t8[['_value_node_network_transmit_bytes']].copy()
        dataset_df['_value_node_network_receive_bytes'] = t1_t9[['_value_node_cpu_user_usage_avg']].copy()
        dataset_df['_value_node_network_transmit_bytes'] = t1_t10[['_value_node_cpu_sys_usage_avg']].copy()
        dataset_df['_value_node_disk_read_bytes'] = t1_t11[['_value_node_memory_usage']].copy()
        dataset_df['_value_node_disk_write_bytes'] = t1_t12[['_value_node_disk_read_bytes']].copy()
        dataset_df['_value_node_disk_write_bytes'] = t1_t13[['_value_node_disk_write_bytes']].copy()
    else:
        df = t1_t2[['_value_container_cpu_average', '_value_container_memory_usage', 'comm_container_cpu_average',
                    'device_id_container_cpu_average']].copy()
        df['_value_node_mem_usage_average'] = t1_t3[['_value_container_disk_read_bytes']].copy()
        df['_value_pod_mem_usage_average'] = t1_t4[['_value_container_disk_write_bytes']].copy()
        df['_value_pod_disk_read_bytes'] = t1_t5[['_value_container_network_receive_bytes']].copy()
        df['_value_pod_disk_write_bytes'] = t1_t6[['_value_container_network_transmit_bytes']].copy()
        df['_value_pod_network_receive_bytes'] = t1_t7[['_value_node_network_receive_bytes']].copy()
        df['_value_pod_network_transmit_bytes'] = t1_t8[['_value_node_network_transmit_bytes']].copy()
        df['_value_node_network_receive_bytes'] = t1_t9[['_value_node_cpu_user_usage_avg']].copy()
        df['_value_node_network_transmit_bytes'] = t1_t10[['_value_node_cpu_sys_usage_avg']].copy()
        df['_value_node_disk_read_bytes'] = t1_t11[['_value_node_memory_usage']].copy()
        df['_value_node_disk_write_bytes'] = t1_t12[['_value_node_disk_read_bytes']].copy()
        df['_value_node_disk_write_bytes'] = t1_t13[['_value_node_disk_write_bytes']].copy()

        dataset_df = pd.concat([dataset_df, df], axis=0, ignore_index=True)

print(dataset_df.shape)
dataset_df.to_csv("dataset/monitor_data/edges/edge_new_dataset.csv")

print(comm_delay_df.shape)
comm_delay_df.to_csv("dataset/monitor_data/edges/edge_comm_delay_dataset.csv")

print(comm_delay_df.shape)
deplay_delay_avg_df.to_csv("dataset/monitor_data/edges/edge_epoch_delay_avg_dataset.csv")