"""
    2 layers CNN for Cifar-10 task

    Warnings: None

    Author: Rui Sun

"""

import torch.nn as nn
import torch.nn.functional as F

class CNN_L2(nn.Module):
    def __init__(self):
        super(CNN_L2, self).__init__()
        # 32 x 32 x 3
        self.conv1 = nn.Conv2d(in_channels=3, out_channels=16, kernel_size=3, padding=1)
        # 16 x 16 x 16
        self.conv2 = nn.Conv2d(in_channels=16, out_channels=32, kernel_size=3, padding=1)
        # 8x8x32
        self.conv3 = nn.Conv2d(in_channels=32, out_channels=64, kernel_size=3, padding=1)
        self.pool = nn.MaxPool2d(2, 2)
        # linear layer (64 * 4 * 4 -> 500)
        self.fc1 = nn.Linear(64 * 4 * 4, 500)
        # linear layer (500 -> 10)
        self.fc2 = nn.Linear(500, 10)
        # dropout层 (p=0.3)
        self.dropout = nn.Dropout(0.3)

    def forward(self, inputs):
        # add sequence of convolutional and max pooling layers
        x = self.pool(F.relu(self.conv1(inputs)))
        x = self.pool(F.relu(self.conv2(x)))
        x = self.pool(F.relu(self.conv3(x)))
        # flatten image input
        x = x.view(-1, 64 * 4 * 4)
        # add dropout layer
        x = self.dropout(x)
        # add 1st hidden layer, with relu activation function
        x = F.relu(self.fc1(x))
        # add dropout layer
        x = self.dropout(x)
        # add 2nd hidden layer, with relu activation function
        x = self.fc2(x)
        return x
