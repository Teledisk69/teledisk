#!/bin/bash

echo sudo apt-get update
echo sudo apt-get upgrade
echo sudo apt-get install make git zlib1g-dev libssl-dev gperf cmake g++
echo git clone --recursive https://github.com/tdlib/telegram-bot-api.git
echo cd telegram-bot-api
echo rm -rf build
echo mkdir build
echo cd build
echo cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=.. ..
echo cmake --build . --target install
echo cd ../..
echo ls -l telegram-bot-api/bin/telegram-bot-api*
