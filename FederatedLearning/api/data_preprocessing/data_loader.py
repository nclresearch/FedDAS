"""
    Utils Data

    Warnings: None

    Author: Rui Sun
"""
from api.data_preprocessing.cifar10.data_loader import load_data

def get_data_loader(dataset_name, cids, args):

    global_data = None
    if dataset_name == "cifar10":
        train_data_num, test_data_num, train_data_global, test_data_global, train_data_local_num_dict, train_data_local_dict, test_data_local_dict, class_num = load_data(cids, args)
        global_data = {"train_data_num": train_data_num, "test_data_num" :test_data_num, "train_data_global": train_data_global,
                       "test_data_global": test_data_global, "class_num": class_num}

    return global_data
