#!/bin/bash

colonies fs sync -l src -d src --yes
colonies fs sync -l data -d data --yes 
colonies fs sync -l result -d result --keeplocal=false --yes
colonies fs sync -l classifier_results -d classifier_results --keeplocal=false --yes
