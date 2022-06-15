'''
    MQTT Manager

    Author: Rui Sun
'''

from loguru import logger

import sys,os
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from core.communication.mqtt.MQTTClient import MQTTClient
except ImportError:
    from UbiFL.core.communication.mqtt.MQTTClient import MQTTClient

class MQTTClientManager(object):

    def __init__(self):
        self.all_mqtt_clients = {}

    def register(self, client_id):
        if self.__is_client_exist(client_id):
            logger.info(f"{client_id} already exist!")
            return None
        client = MQTTClient(client_id=client_id)
        self.all_mqtt_clients[client_id] = client
        return client

    def get_client(self,client_id):
        if not self.__is_client_exist(client_id):
            logger.error(f"{client_id} is not exist!")
            return None
        return self.all_mqtt_clients.get(client_id)

    def __is_client_exist(self, client_id):
        return client_id in self.all_mqtt_clients.keys()

    def del_client(self,client_id):
        if not self.__is_client_exist(client_id):
            logger.error(f"{client_id} is not exist!")
            return False
        self.all_mqtt_clients.pop(client_id)
        return True
