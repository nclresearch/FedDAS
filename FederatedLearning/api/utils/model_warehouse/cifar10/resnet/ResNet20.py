"""
    20 layers RestNet for Cifar-10 task

    Warnings: None

    Author: Rui Sun

"""

import torch
import torch.nn as nn
import torch.nn.functional as F

class ResBlock(nn.Module):
    def __init__(self, ch_in, ch_out, stride=1):
        super(ResBlock, self).__init__()

        self.conv1 = nn.Conv2d(ch_in, ch_out, kernel_size=3, stride=stride, padding=1)
        self.bn1 = nn.BatchNorm2d(ch_out)
        self.conv2 = nn.Conv2d(ch_out, ch_out, kernel_size=3, stride=1, padding=1)
        self.bn2 = nn.BatchNorm2d(ch_out)

        self.extra = nn.Sequential()
        if ch_out != ch_in:
            self.extra = nn.Sequential(
                nn.Conv2d(ch_in, ch_out, kernel_size=1, stride=stride),
                nn.BatchNorm2d(ch_out)
            )

    def forward(self, x):
        out = self.conv1(x)
        out = self.bn1(out)
        out = F.relu(out)
        out = self.conv2(out)
        out = self.bn2(out)
        out = self.extra(x) + out
        out = F.relu(out)
        return out

class ResNet20(nn.Module):

    def __init__(self):
        super(ResNet, self).__init__()

        self.conv1 = nn.Sequential(
            nn.Conv2d(3, 64, kernel_size=3, stride=3, padding=0),
            nn.BatchNorm2d(64)
        )

        # self.block1 = ResBlock(64, 128, stride=2)
        # self.fc = nn.Linear(128 * 1 * 1, 10)

        self.block1 = ResBlock(64, 128, stride=2)
        self.block2 = ResBlock(128, 256, stride=2)
        self.block2 = ResBlock(128, 256, stride=2)
        self.block3 = ResBlock(256, 512, stride=2)
        self.block4 = ResBlock(512, 512, stride=2)
        self.fc = nn.Linear(512 * 1 * 1, 10)

    def forward(self, x):
        x = self.conv1(x)
        x = F.relu(x)
        x = self.block1(x)
        x = self.block2(x)
        x = self.block3(x)
        x = self.block4(x)
        x = F.adaptive_avg_pool2d(x, [1, 1])
        x = torch.flatten(x, 1)
        out = self.fc(x)
        return out