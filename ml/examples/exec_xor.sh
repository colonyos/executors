#!/bin/bash

snapshotid=$(openssl rand -hex 10) 
colonies fs sync -l src -d src --yes
colonies fs snapshot create -l src -n $snapshotid 
colonies function exec --func execute --kwargs cmd:python3,args:/tmp/xor/xor.py --snapshots $snapshotid:/tmp/xor --targettype gpu-mlexecutor --follow
colonies fs snapshot remove -n $snapshotid


