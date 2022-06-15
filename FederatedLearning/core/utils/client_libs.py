"""
    Fl lib for executor

    Warnings: None

    Author: Rui Sun
"""
import sys,os
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../")))
sys.path.insert(0, os.path.abspath(os.path.join(os.getcwd(), "../../../UbiFL")))
try:
    from core.utils.fllibs import *
    from core.utils.arguments import args
except ImportError:
    from UbiFL.core.utils.fllibs import *
    from UbiFL.core.utils.arguments import args
