#!/bin/bash
set -e

git clone https://github.com/lbradstreet/WFG-hypervolume.git
cd WFG-hypervolume
    cp ../avl.h ../dict_info.h .
    make march=x86-64 # insert the appropriate architecture name
cd ..
cp WFG-hypervolume/wfg0 .
cp WFG-hypervolume/wfg2 .
rm -rf WFG-hypervolume
