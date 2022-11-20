#!/bin/bash

echo "build with gcc"
g++ $(< flags_gcc) -o dnsperf_gcc dnsperf_main.cpp dnsquery.cpp utils.cpp -lunbound -lleveldb

