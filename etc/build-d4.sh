#!/bin/bash
set -e

git clone https://github.com/crillab/d4.git d4-repository
cd d4-repository
    make -j8
cd ..
cp d4-repository/d4 .
rm -rf d4-repository
