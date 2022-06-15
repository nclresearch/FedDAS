"""
    Data Loader for cifar10

    Warnings: None

    Author: Rui Sun
"""
import torch
# from fedml.data.cifar10.data_loader import load_partition_data_cifar10

import sys,os
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../../UbiFL")))
try:
    from core.utils.general import test_mkdir
    from api.data_preprocessing.fedml_data.cifar10.data_loader import load_partition_data_cifar10
except ImportError:
    from UbiFL.core.utils.general import test_mkdir
    from UbiFL.api.data_preprocessing.fedml_data.cifar10.data_loader import load_partition_data_cifar10


def load_data(cids,args):

    dataset = None

    if args.partition_method == "homo":
        train_data_num, test_data_num, train_data_global, test_data_global, \
        train_data_local_num_dict, train_data_local_dict, test_data_local_dict, \
        class_num = load_partition_data_cifar10(args.dataset, args.data_dir, args.partition_method,
                                args.partition_alpha, args.num_of_clients, args.batch_size)

        for k,v in train_data_local_dict.items():
            save_dataset_as_file(cids[k], v.dataset, False)
            save_dataset_as_file(cids[k], test_data_local_dict[k].dataset, True)

        dataset = [train_data_num, test_data_num, train_data_global, test_data_global,
         train_data_local_num_dict, train_data_local_dict, test_data_local_dict, class_num]

    return dataset

def save_dataset_as_file(cid, data_loader, is_test):

    if is_test:
        dir_path = os.path.join("cache", "test_data", "partition")
        data_saving_path = os.path.join(dir_path, "test-" + cid + ".pth")
    else:
        dir_path = os.path.join("cache", "train_data", "partition")
        data_saving_path = os.path.join(dir_path, "train-" + cid + ".pth")

    test_mkdir(dir_path)
    torch.save(data_loader,data_saving_path)
