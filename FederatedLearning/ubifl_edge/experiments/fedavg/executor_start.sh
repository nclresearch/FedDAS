#!/usr/bin/env bash

CONF_PATH=$1

WORKING_PATH=$2

wandb login 49094e75af25687267b35f6567d108f0084b4605
wandb online

#wandb off

python3 $UBIFL_HOME/ubifl_edge/experiments/fedavg/executor.py -cf $CONF_PATH -wp $WORKING_PATH