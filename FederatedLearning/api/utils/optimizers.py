"""
    Optimizers

    Warnings: None

    Author: Rui Sun
"""

from torch import optim


class Optimizers:
    def __init__(self, params, lr):
        self.params = params
        self.lr = lr

    def get_adam_optimizer(self):
        return optim.Adam(params=self.params, lr=self.lr)

    def get_sgd_optimizer(self):
        return optim.SGD(params=self.params, lr=self.lr)

class ServerOptimizers(Optimizers):

    def __init__(self, params, lr):
        super(ServerOptimizers, self).__init__(params, lr)


class ClientOptimizers(Optimizers):

    def __init__(self, params, lr):
        super(ClientOptimizers, self).__init__(params, lr)
