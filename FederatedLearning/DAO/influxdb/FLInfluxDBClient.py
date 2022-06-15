"""
    InfluxDB mqtt_client

    Warnings: None

    Author: Rui Sun

"""
import time

from influxdb_client import InfluxDBClient
from influxdb_client.client.write_api import SYNCHRONOUS

class FLInfluxDBClient(InfluxDBClient):

    def __init__(self, token, org, url=None, timeout=1_000_000, ip='localhost', port='8086', **kwargs):

        # ============= Basic Info =============
        if url is None:
            url = "http://" + ip + ":" + str(port)

        super(FLInfluxDBClient, self).__init__(url=url, token=token, timeout=timeout, org=org, **kwargs)
        self.ip = ip
        self.port = port


    def connect(self):
        self.__write_api = super(FLInfluxDBClient, self).write_api(write_options=SYNCHRONOUS)

    def write_new_records(self, bucket, data):
        '''
        Write a new record to DAO
        :param bucket: Target bucket name
        :param data:
                1. if it is a str, will insert one record
                2. if it is an array, will insert a batch records
                3. if it also could be a Point type implement from influxdb_client import Point
        :return: result
            success: None
        '''
        return self.__write_api.write(bucket, self.org, data)

    def query(self, expr):
        return super(FLInfluxDBClient, self).query_api().query_data_frame(expr, org=self.org)

    def query_stream(self, expr):
        '''
            Stream result,

            large_stream = query_api().query_stream(expr, org=self.org)

            for record in large_stream:
                if record["location"] == "New York":
                    break

            large_stream.close()
        '''
        return super(FLInfluxDBClient, self).query_api().query_stream(expr, org=self.org)

    # def close(self):
    #     super(FLInfluxDBClient, self).close()

if __name__ == '__main__':
    client = FLInfluxDBClient(token="TokenForNCLProject",org="NCL",timeout=5000,ip="10.70.31.211",port="12186")

    # client.connect()
    # client.write_new_records("pre_setting","server_running_status,device_id=1,comm=1 comm_delay=0.1")
    #
    # client.write_new_records("pre_setting", "server_running_status,device_id=1,comm=1 comm_delay=0.2")

    res = client.query_api().query('q=SHOW SERIES')
    print(res)

    # c = InfluxDBClient("pre_setting",token="TokenForNCLProject",org="NCL",timeout=5000,url="http://10.79.253.130:30004")

    # time.sleep(5)
    # client.close()