"""
    Client Manager

    Warnings: None

    Author: Rui Sun
"""
import uuid
import shortuuid
from loguru import logger


class ClientManager(object):

    def __init__(self, client_sample_mode, args):
        # >>>>>>>>>>> Origin FedAVG Conf >>>>>>>>>>>>
        self.data_set_name = args.dataset
        self.is_iid = args.is_iid
        self.num_of_clients = args.num_of_clients
        self.selected_num_of_clients = int(max((args.num_of_clients * args.client_fraction), 1))
        self.clients_set = {}

        # Client management
        self.finished_training_clients = []
        self.feasible_clients = []
        self.registered_clients = []

        # <<<<<<<<<<< Origin FedAVG Conf <<<<<<<<<<<<

        # TODO mqtt_client sample mode to finish this for mqtt_client selection
        self.client_sample_mode = client_sample_mode
        self.args = args

    def register_client(self, hostId):
        """
        Register an new mqtt_client
        :param hostId:
        :return:
        """
        cid = self.get_short_uuid_4bit(hostId)

        self.registered_clients.append(cid)

        return cid

    def client_quit(self, clientID):
        if clientID not in self.feasible_clients:
            logger.warning(f'No mqtt_client {clientID} in the list, remove failed !')
            return 2
        self.feasible_clients.remove(clientID)

    def client_join(self, clientID):
        if clientID not in self.feasible_clients:
            self.feasible_clients.append(clientID)

    def get_all_feasible_clients(self):
        return self.feasible_clients

    def get_all_registered_clients(self):
        return self.registered_clients

    def get_all_clients_len(self):
        return len(self.feasible_clients)

    def get_uuid_32bit(self, hostId):
        return str(hostId) + "_" + uuid.uuid4().hex

    def get_short_uuid_4bit(self, hostId):
        return str(hostId) + "_" + str(shortuuid.ShortUUID().random(length=4))

    def clear_finished_clients_queue(self):
        self.finished_training_clients.clear()
