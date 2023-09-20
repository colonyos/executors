#!/bin/bash

colonies fs sync -l src -d src --yes
colonies fs sync -l data -d data --yes 
colonies fs sync -l results -d results --keeplocal=false --yes
