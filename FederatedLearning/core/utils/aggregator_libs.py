"""
    Fl lib for aggregator

    Warnings: None

    Author: Rui Sun
"""
import sys,os
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../UbiFL")))
try:
    from core.utils.general import init_logger
    from core.utils.fllibs import *
    from core.utils.arguments import args
except ImportError:
    from UbiFL.core.utils.general import init_logger
    from UbiFL.core.utils.fllibs import *
    from UbiFL.core.utils.arguments import args



def load_aggregator_settings():
    """
    Load all initial function for aggregator
    :return:
    """
    log_dir = os.path.join(args.log_path, args.time_stamp, 'aggregator')
    log_file = os.path.join(log_dir, 'log')
    init_logger(log_dir, log_file)
