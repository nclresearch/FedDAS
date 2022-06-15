# Ubiquitous Federated Learning(UbiFL)
UbiFL is designed for Edge-IoT computing in large scale, includes auto-deployment, distributed system monitoring, fault detection and diagnosis.


## System Environments

### Python Dependent Packages
|   Packagename   | Version | Note |
|:---------------:|:-------:| :---: |
|     loguru      |  0.5.3  | == or latest |
|      tqdm       | 4.62.3  | == or latest |
|     pyyaml      |   6.0   | == or latest |
|    paho-mqtt    |  1.6.1  | == or latest |
|      flask      |  2.0.2  | == or latest |
|   matplotlib    |   3.5   | == or latest |
|      wandb      | 0.12.7  | == or latest |
| influxdb-client | 1.24.0  | == or latest |
|     pandas      |  1.4.2  | == or latest |



### Directory structure (manually create required)
UbiFL

|-- cache

|   |-- accuracy

|   |-- test_data

|   |   |-- partition

|   |-- train_data

|   |   |-- partition

|-- running

|   |-- checkpoints

|   |-- logs


### Framework Startup Parameters 
| Abbreviated Parameter Name | Full Parameter Name | Description | Default Value | Tips |
| :---: | :---: | :---: | :---: | :---: |
| -g | --gpu | Id of GPU that want to use | 0 | |

### MQTT Topic Parttern
**aggregator subscribe/executor publish:** /client/publish/ + cid
**aggregator publish/executor subscribe:** /aggregator/publish/ + cid

### Data Partition
**Train data & labels:** cache/train_data/partition , File example: cache/train_data/partition/train-XXXXXXXX.pth

**Test data & labels:** cache/test_data/partition , File example: cache/test_data/partition/test-XXXXXXXX.pth

## Framework Running Environment

### Pytorch Install
**Linux install pytorch 1.7 + CUDA 10.1**
```shell
pip3 install torch==1.7.1+cu101 torchvision==0.8.2+cu101 torchaudio==0.7.2 -f https://download.pytorch.org/whl/torch_stable.html
```

**Linux install pytorch 1.9.1 + CUDA 11.1**
```shell
pip3 install torch==1.9.1+cu111 torchvision==0.10.1+cu111 torchaudio==0.9.1 -f https://download.pytorch.org/whl/torch_stable.html
```

**Linux install pytorch 1.10.0 + CUDA 11.3**
```shell
pip3 install torch==1.10.0+cu113 torchvision==0.11.1+cu113 torchaudio==0.10.0+cu113 -f https://download.pytorch.org/whl/cu113/torch_stable.html
```

**Windows install pytorch 1.9 + CUDA 11.1**
```shell
pip3 install torch==1.9.0+cu111 torchvision==0.10.0+cu111 torchaudio==0.9.0 -f https://download.pytorch.org/whl/torch_stable.html
```

### Docker install
#### Ubuntu
ref: [[Docker Webpage](https://docs.docker.com/engine/install/ubuntu/)]

**Uninstall old versions**
```shell
 sudo apt-get remove docker docker-engine docker.io containerd runc
```

**Set up the repository**
```shell
# Update the apt package index and install packages to allow apt to use a repository over HTTPS:
 sudo apt-get update
 sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release
    
# Add Docker’s official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

#Use the following command to set up the stable repository. To add the nightly or test repository, add the word nightly or test (or both) after the word stable
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

**Install Docker Engine**
```shell
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io
```

### Message Queue - MQTT

MQTT installation via docker

```shell
sudo docker pull registry.cn-hangzhou.aliyuncs.com/synbop/emqttd:2.3.6

sudo docker run --name emq -p 18083:18083 -p 1883:1883 -p 8084:8084 -p 8883:8883 -p 8083:8083 -d registry.cn-hangzhou.aliyuncs.com/synbop/emqttd:2.3.6
```

**Modify the maximum message size (large model parameters):**

1. vi emqttd/etc/emq.conf
2. To find mqtt.max_packet_size change to 100MB (Or message_size_limit 0
)
3. Restart container

### Monitoring
#### Influxdb 2.1.1 (via docker)
```shell
sudo docker pull influxdb

sudo docker run \
      -d -p 18883:8083 \
      -p 8086:8086 \
      -v $PWD/influxdb2/data:/var/lib/influxdb2 \
      -v $PWD/influxdb2/config:/etc/influxdb2 \
      -e DOCKER_INFLUXDB_INIT_MODE=setup \
      -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
      -e DOCKER_INFLUXDB_INIT_PASSWORD=admin123 \
      -e DOCKER_INFLUXDB_INIT_ORG=NCL \
      -e DOCKER_INFLUXDB_INIT_BUCKET=pre_setting \
      -e DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=TokenForNCLCIKM2022DemoPaperProject \
      --name my_influxdb influxdb
```
##### Databse Setting
**Note: All following org and buckets should be created manually**
**org:** NCL

| Bucket | Description |
| :----: | :---- |
| pre_setting | Record the system presetting like hyper parameter settings |
| running_monitoring | Monitor the running status of model |

### Points
#### Cloud
|      Name     | Value |Description|Related Symptom|
|:----|:------|:---|:---|
|node_cpu_average||||
|pod_cpu_average||||
|node_mem_usage_average||||
|pod_mem_usage_average||||
|pod_disk_read_bytes||||
|pod_disk_write_bytes||||
|pod_network_receive_bytes||||
|pod_network_transmit_bytes||||
|node_network_receive_packets||||
|node_network_receive_dropped_packets||||
|node_network_transmit_dropped_packets||||
|node_network_receive_errors||||
|node_network_transmit_errors||||
|node_network_receive_bytes||||
|node_network_transmit_bytes||||
|node_network_receive_dropped_packets||||
|node_network_transmit_dropped_packets||||
| node_network_receive_errors         ||||
| node_network_transmit_errors        ||||
| node_disk_read_bytes                ||||
| node_disk_write_bytes               ||||

#### Edge
| Name                                       | Value |Description|Related Symptom|
|:-------------------------------------------|:------|:---|:---|
| container_cpu_average                      ||||
| container_memory_usage                     ||||
| container_disk_read_bytes                  ||||
| container_disk_write_bytes                 ||||
| container_network_receive_bytes            ||||
| container_network_transmit_bytes           ||||
| container_network_receive_dropped_packets  ||||
| container_network_transmit_dropped_packets ||||
| container_network_receive_errors           ||||
| container_network_transmit_errors          ||||
| node_network_receive_bytes                 ||||
| node_network_transmit_bytes                ||||
| node_network_receive_dropped_packets       ||||
| node_network_transmit_dropped_packets      ||||
| node_network_receive_errors                ||||
| node_network_transmit_errors               ||||
| node_disk_read_bytes                       ||||
| node_disk_write_bytes                      ||||

## Notes:
If your wanna monitor Jetson Nano, you have to install jtop in your system first.
