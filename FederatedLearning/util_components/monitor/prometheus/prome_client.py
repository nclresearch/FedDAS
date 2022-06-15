"""

    Warnings: None

    Author: Rui Sun
"""
import json
import urllib.parse

import requests
from loguru import logger

class PrometheusClient:
    def __init__(self,ip,port):
        self.port = port
        self.ip = ip
        self.http_addr = f"http://{ip}:{port}"

        # Limit times to retry getting result
        # Warnings some problems when calling these two parameters if they didn't claim in the super class
        self.retry_times = 20
        self.retry_fake_value = None

    def check_target(self):
        """
        Check all of services status
        :return:
            alive_list: the list of services who are still alive , eg,. 10.79.253.130:6443 or 19scomps007
            down_list: the list of services who are in down starus, eg,. 10.79.253.130:6443 or 19scomps007
        """
        url = f"{self.http_addr}/api/v1/targets"
        response = requests.request('GET', url)

        alive_list = []
        down_list = []
        if response.status_code == 200:
            targets = response.json()['data']['activeTargets']
            alive_num, total_num = 0, 0
            for target in targets:
                total_num += 1
                if target['health'] == 'up':
                    alive_num += 1
                    alive_list.append(target['labels']['instance'])
                else:
                    down_list.append(target['labels']['instance'])
            logger.info("-----------------------TargetsStatus--------------------------")

            logger.info(f"{alive_num} in {total_num} Targets are alive !!!")
            logger.info("--------------------------------------------------------------")
            for down in down_list:
                logger.warning(f"{down} down !!!")
            logger.info("-----------------------TargetsStatus--------------------------")
        else:
            logger.error("Get targets status failed!")

        return alive_list, down_list

    def query_by_expr(self,expr):
        url = f"{self.http_addr}/api/v1/query?query={urllib.parse.quote(expr)}"
        return json.loads(requests.get(url=url).content.decode('utf8', 'ignore'))

    def extract_result(self,resp):
        '''
        :param resp: now support one result
        :return:
        '''

        if 'data' in resp.keys():
            if 'result' in resp['data'].keys():
                if len(resp['data']['result']) != 0:
                    return resp['data']['result']
        return None

