#!/bin/sh

set -x
influx bucket create -n running_monitoring -r 30d
