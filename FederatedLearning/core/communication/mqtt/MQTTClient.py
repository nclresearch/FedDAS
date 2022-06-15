'''
    Library for MQTT
    Author: Rui Sun
'''

import sys
import pickle

from loguru import logger
from paho.mqtt import client


class MQTTClient(client.Client):

    def __init__(self, client_id, clean_session=True):
        self.fl_client_id = client_id
        super(MQTTClient, self).__init__(clean_session=clean_session)

    def set_subscribe_topic(self, subscribe_topic):
        self.subscribe_topic = subscribe_topic

    def set_publish_topic(self, publish_topic):
        self.publish_topic = publish_topic

    def connect(self, host, port=1883, keepalive=65535, bind_address="", bind_port=0,
                clean_start=3, properties=None):
        # For MQTT V5, use the clean start flag only on the first successful connect
        # clean_start = MQTT_CLEAN_START_FIRST_ONLY = 3

        # Callback for mqtt connection
        def on_connect(mqtt_client, userdata, flags, rc):
            if rc != 0:
                logger.error(f'Failed to connect, return code {rc}.')

        # Set Connecting Client ID
        self.on_connect = on_connect

        super().connect(host=host, port=port, keepalive=keepalive, bind_address=bind_address, bind_port=bind_port,
                clean_start=clean_start, properties=properties)

    def subscribe(self, topic, qos=0, options=None, properties=None):
        try:
            super().subscribe(topic=topic, qos=qos, options=options, properties=properties)
        except Exception:
            logger.error(f'The {self.fl_client_id} subscribe MQTT topic: {topic} is failed !')

        logger.info(f'Subscribed server to mqtt_client topic: {topic}')

    def publish(self, topic, payload=None, qos=0, retain=False, properties=None):

        payload = pickle.dumps(payload)

        self.get_payload_size(payload)

        res = super().publish(topic=topic, payload=payload, qos=qos, retain=retain, properties=properties)
        # 等待消息成功发送, 此处肯能有通讯问题，需要配合qos使用
        # res.wait_for_publish()

        if res[0] != 0:
            logger.error(f'The mqtt_client {self.fl_client_id} send message: {payload}, to {topic} is failed !')
        else:
            logger.info(f'{self.fl_client_id} sent msg to topic: {topic}')

    def get_payload_size(self, payload):
        # Monitor size of model
        if int(sys.getsizeof(payload) / 1024) == 0:
            logger.info(
                f'Payload size is : {int(sys.getsizeof(payload))} b')
        elif int(sys.getsizeof(payload) / 1024 / 1024) == 0:
            logger.info(
                f'Payload size is : {int(sys.getsizeof(payload) / 1024)} Kb')
        elif int(sys.getsizeof(payload) / 1024 / 1024 / 1024) == 0:
            logger.warning(
                f'Payload size is : {int(sys.getsizeof(payload) / 1024 / 1024)} Mb')
        else:
            logger.warning(
                f'Payload size is : {int(sys.getsizeof(payload) / 1024 / 1024 / 1024)} Gb')