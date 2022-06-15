from MQTTClientManager import MQTTClientManager

if __name__ == '__main__':
    manager = MQTTClientManager()
    mq_client = manager.register("1")
    mq_client.connect( host= "localhost",port=1883)
    mq_client.publish("/aggregator/publish/123",payload="hello")
    mq_client.mqtt_client.loop_start()
