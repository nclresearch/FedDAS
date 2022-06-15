"""
    JTop monitor for jetson nano

    Warnings: None

    Author: Rui Sun
"""

from jtop import jtop

class JTopMonitor(object):
    '''
        Backup for override
    '''

    def __init__(self):
        self.jetson = jtop()

    def start(self):
        '''
        Get running status, example result:
            {'time': datetime.datetime(2022, 5, 30, 16, 37, 20, 547814),
            'uptime': datetime.timedelta(33, 22246, 370000), 'jetson_clocks': 'OFF', 'nvp model': 'MAXN', 'CPU1': 41,
            'CPU2': 54, 'CPU3': 61, 'CPU4': 48, 'GPU': 99, 'RAM': 3656268, 'EMC': 3656268, 'IRAM': 3656268, 'SWAP': 1982,
            'APE': 25, 'NVENC': 'OFF', 'NVDEC': 'OFF', 'NVJPG': 'OFF', 'fan': 31.372549019607842, 'Temp AO': 50.5,
            'Temp CPU': 44.0, 'Temp GPU': 39.5, 'Temp PLL': 37.5, 'Temp iwlwifi': 40.0, 'Temp thermal': 41.5,
            'power cur': 4664, 'power avg': 2852}
        '''
        self.jetson.start()

    @property
    def stats(self):
        return self.jetson.stats

    def close(self):
        self.jetson.close()

    def ok(self):
        return self.jetson.ok()

    @property
    def board(self):
        '''
        Get board info Example result:
            {'info': {'machine': 'NVIDIA Jetson Nano (Developer Kit Version)', 'jetpack':
            '4.6', 'L4T': '32.6.1'}, 'hardware': {'TYPE': 'Nano (Developer Kit Version)', 'CODENAME': 'porg',
            'SOC': 'tegra210', 'CHIP_ID': '33', 'BOARDIDS': '3448', 'MODULE': 'P3448-0000', 'BOARD': 'P3449-0000',
            'CUDA_ARCH_BIN': '5.3', 'SERIAL_NUMBER': '1422019082608'}, 'libraries': {'CUDA': '10.2.300',
            'cuDNN': '8.2.1.32', 'TensorRT': '8.0.1.6', 'VisionWorks': '1.6.0.501', 'OpenCV': '4.1.1', 'OpenCV-Cuda':
            'NO', 'VPI': 'ii libnvvpi1 1.1.15 arm64 NVIDIA Vision Programming Interface library', 'Vulkan': '1.2.70'}}
        '''
        return self.jetson.board
