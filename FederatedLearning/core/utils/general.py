"""
    System lib

    Warnings: None

    Author: Rui Sun

"""
import os

import yaml
from loguru import logger

def init_logger(log_dir,log_file_name):
    """
    Logger setting
    :return:
    """
    if not os.path.isdir(log_dir):
        os.makedirs(log_dir, exist_ok=True)

    logger.add(log_file_name, format="{time}|{level}|{message}", level="INFO", rotation="10 MB")

def load_yaml_conf(yaml_file):
    with open(yaml_file) as fin:
        data = yaml.load(fin, Loader=yaml.FullLoader)
    return data

def test_mkdir(path):

    if not os.path.isdir(path):
        # logger.info(path)
        # os.mkdir(path)
        os.makedirs(path)

def load_yaml_conf(yaml_file):
    with open(yaml_file) as fin:
        data = yaml.load(fin, Loader=yaml.FullLoader)
    return data