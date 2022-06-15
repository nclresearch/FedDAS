from hw_nas_bench_api import HWNASBenchAPI as HWAPI
hw_api = HWAPI("../HW-NAS-Bench-v1_0.pickle", search_space="nasbench201")

# Example to get all the hardware metrics in the No.0,1,2 architectures under NAS-Bench-201's Space
for idx in range(3):
    for dataset in ["cifar10", "cifar100", "ImageNet16-120"]:
        HW_metrics = hw_api.query_by_index(idx, dataset)
        print("The HW_metrics (type: {}) for No.{} @ {} under NAS-Bench-201: {}".format(type(HW_metrics),
                                                                               idx,
                                                                               dataset,
                                                                               HW_metrics))



config = hw_api.get_net_config(2, "cifar10")
print(config)
from hw_nas_bench_api.nas_201_models import get_cell_based_tiny_net
network = get_cell_based_tiny_net(config) # create the network from configurration
print(network) # show the structure of this architecture




# The index in FBNet Space is not a number but a list with 22 elements, and each element is from 0~8
# from hw_nas_bench_api import HWNASBenchAPI as HWAPI
# hw_api = HWAPI("../HW-NAS-Bench-v1_0.pickle", search_space="fbnet")
#
# # Example to get all the hardware metrics in 3 specfic architectures under FBNet's Space
# print("===> Example to get all the hardware metrics in the No.0,1,2 architectures under FBNet's Space")
# for idx in [[0]*22, [0]*21+[1]*1, [0]*20+[1]*2]:
#     for dataset in ["cifar100", "ImageNet"]:
#         HW_metrics = hw_api.query_by_index(idx, dataset)
#         print("The HW_metrics (type: {}) for No.{} @ {} under NAS-Bench-201: {}".format(type(HW_metrics),
#                                                                                idx,
#                                                                                dataset,
#                                                                                HW_metrics))