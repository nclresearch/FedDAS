from monitoring.anomaly_detector import FLInfluxDBClient

influx_client = FLInfluxDBClient(token="TokenForNCLWWW2022DemoPaperProject", org='NCL', ip="10.79.253.130", port="30004")

start_time = '2022-01-22T21:43:36Z'
stop_time = '2022-01-23T11:45:00Z'

expr = [
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_cpu_usage") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, pod_cpu_usage: t2}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t3 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_mem_usage_average") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, node_mem_usage_average: t3}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t4 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_mem_usage_average") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, pod_mem_usage_average: t4}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t5 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, pod_disk_read_bytes: t5}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t6 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, pod_disk_write_bytes: t6}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t7 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, pod_network_receive_bytes: t7}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t8 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, pod_network_transmit_bytes: t8}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t9 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_network_receive_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, node_network_receive_bytes: t9}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t10 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_network_transmit_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, node_network_transmit_bytes: t10}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t11 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_disk_read_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, node_disk_read_bytes: t11}, on: ["_time"])',
'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_cpu_average") |> sort(columns:["_time"]) t12 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "node_disk_write_bytes") |> sort(columns:["_time"]) join(tables: {node_cpu_average: t1, node_disk_write_bytes: t12}, on: ["_time"])'
]
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

# query comm delay
expr_comm_delay = 'start_time='+start_time+' stop_time='+stop_time+' t1 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field == "pod_cpu_usage") |> sort(columns:["_time"]) t2 = from(bucket: "running_monitoring") |> range(start: start_time, stop:stop_time) |> filter(fn: (r) => r._measurement == "server_running_status" and r._field  == "comm_delay") |> sort(columns:["_time"])  join(tables: {container_cpu_average: t1, comm_delay: t2}, on: ["comm"])'

comm_delay_df = influx_client.query(expr_comm_delay)

dataset_df = t1_t2[['_value_node_cpu_average','_value_pod_cpu_usage','comm_node_cpu_average','device_id_node_cpu_average']].copy()
dataset_df['_value_node_mem_usage_average'] = t1_t3[['_value_node_mem_usage_average']].copy()
dataset_df['_value_pod_mem_usage_average'] = t1_t4[['_value_pod_mem_usage_average']].copy()
dataset_df['_value_pod_disk_read_bytes'] = t1_t5[['_value_pod_disk_read_bytes']].copy()
dataset_df['_value_pod_disk_write_bytes'] = t1_t6[['_value_pod_disk_write_bytes']].copy()
dataset_df['_value_pod_network_receive_bytes'] = t1_t7[['_value_pod_network_receive_bytes']].copy()
dataset_df['_value_pod_network_transmit_bytes'] = t1_t8[['_value_pod_network_transmit_bytes']].copy()
dataset_df['_value_node_network_receive_bytes'] = t1_t9[['_value_node_network_receive_bytes']].copy()
dataset_df['_value_node_network_transmit_bytes'] = t1_t10[['_value_node_network_transmit_bytes']].copy()
dataset_df['_value_node_disk_read_bytes'] = t1_t11[['_value_node_disk_read_bytes']].copy()
dataset_df['_value_node_disk_write_bytes'] = t1_t12[['_value_node_disk_write_bytes']].copy()

# df = pd.concat([t1_t2_df,t2_t3_df],axis=0)
print(dataset_df.shape)

print(comm_delay_df.shape)

dataset_df.to_csv("dataset/monitor_data/cloud/cloud_new_dataset.csv")

comm_delay_df.to_csv("dataset/monitor_data/cloud/cloud_comm_delay_dataset.csv")