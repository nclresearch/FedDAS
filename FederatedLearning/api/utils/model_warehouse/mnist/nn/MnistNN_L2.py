"""
    2 layers NN for handwriting classification task

    Warnings: None

    Author: Rui Sun

"""

import torch.nn as nn
import torch.nn.functional as F


class MnistNN_L2(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 200)
        self.fc2 = nn.Linear(200, 200)
        self.fc3 = nn.Linear(200, 10)

    def forward(self, inputs):
        tensor = F.relu(self.fc1(inputs))
        tensor = F.relu(self.fc2(tensor))
        tensor = self.fc3(tensor)
        return tensor