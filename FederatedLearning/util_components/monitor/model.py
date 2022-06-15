"""
    Model monitoring

    Warnings: None

    Author: Rui Sun
"""

from torch.cuda import Event

class GPUEventTimer(Event):
    
    def __init__(self,enable_timing=False, blocking=False, interprocess=False):
        super(GPUEventTimer, self).__init__(enable_timing=enable_timing, blocking=blocking, interprocess=interprocess)

    def record(self, stream=None):
        super(GPUEventTimer, self).record(stream)
        # TODO: send to influxdb and wandb

class CPUTimer(object):
    def __init__(self):
        pass