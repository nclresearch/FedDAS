"""
    Aggregator

    Warnings: None

    Author: Rui Sun
"""
import threading

from flask import Flask, request, send_from_directory

import sys,os
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from api.data_preprocessing.data_loader import get_data_loader
    from api.fedavg.aggregator import Aggregator
    from core.utils.aggregator_libs import *
    from core.utils.arguments import args
except ImportError:
    from UbiFL.api.data_preprocessing.data_loader import get_data_loader
    from UbiFL.api.fedavg.aggregator import Aggregator
    from UbiFL.utils.core.aggregator_libs import *
    from UbiFL.core.utils.arguments import args


app = Flask(__name__)
agg = Aggregator()

'''
    Result code:
    0: fail
    1: success
    2: mqtt_client quit failed because invalid mqtt_client id
    
    Note:
    all mqtt_client should register first, 
    and if they wanna join to training, they should then request join api
'''


@app.route(args.parameter_server_api_client_register, methods=['GET'])
def client_register():
    host_id = request.args.get('hid')

    # Register a mqtt_client in fl mqtt_client manager
    cid = agg.client_manager.register_client(hostId=host_id)

    return cid


@app.route(args.parameter_server_api_client_join, methods=['GET'])
def client_join():
    """
    Client join training
    :return:
    """
    cid = request.args.get('cid')

    if cid not in agg.client_manager.get_all_registered_clients():
        return 'Join failed, invalid cid, register first'

    agg.client_manager.client_join(cid)

    # init mqtt for this mqtt_client
    mqtt_client = agg.mqtt_manager.register(cid)
    # None means this mqtt_client already in training
    if mqtt_client is not None:
        mqtt_client.set_subscribe_topic('/mqtt_client/publish/' + cid)
        mqtt_client.set_publish_topic('/aggregator/publish/' + cid)

    return 'ok'


@app.route(args.parameter_server_api_client_quit, methods=['GET'])
def client_quit():
    cid = request.args.get('cid')

    # clear mqtt_client
    res = agg.client_manager.client_quit(cid)
    if res == 2:
        return 'Quit failed, invalid cid'
    agg.client_manager.client_quit(cid)
    # clear communication
    agg.mqtt_manager.del_client(cid)
    return 'ok'


@app.route(args.parameter_server_api_download_dataset, methods=['GET'])
def download_dataset():
    logger.info(f"Sending dataset {request.args.get('path')} ...")
    return send_from_directory(directory=request.args.get('dir'), path=request.args.get('path'), as_attachment=True)


if __name__ == "__main__":
    setproctitle.setproctitle("UbiFL_Aggregator0")
    t = threading.Thread(target=agg.run)
    t.start()

    app.run(host=args.parameter_server_listen_ip, port=args.parameter_server_listen_port)
