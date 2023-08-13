#!/bin/bash

colonies fs sync -l src -d src
colonies fs sync -l data -d data
colonies fs snapshot create -l src -n srcsnapshot
colonies fs snapshot create -l data -n datasnapshot
