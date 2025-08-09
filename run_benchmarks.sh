#!/usr/bin/env bash

cd /app

sudo chmod 777 /app/output/
mkdir -p /app/output/dnsperfcxx
mkdir -p /app/output/dnsperfgo
mkdir -p /app/output/dnsperfpy
mkdir -p /app/output/dnsperfpycxx

# Run DNS perf queries
# for i in {1..5}; do 
n=5
for ((i=0; i < n; i++)); do
    echo "Starting loop ${i} out of ${n}"
    
    echo "Running C++ version"
    time /app/cxx/dnsperf_gcc  --servers servers.txt --domains domains.txt --output /app/output/dnsperfcxx
    sleep 1

    echo "Running Go version"
    time /app/golang/dnsperf --servers servers.txt --domains domains.txt --output /app/output/dnsperfgo --verbose=false
    sleep 1

    echo "Running Python version"
    time python3 /app/python/pydnsperf_main.py --servers servers.txt --domains domains.txt --output /app/output/dnsperfpy 
    sleep 1

    # echo "Running Python with C++ wrapper version"
    # time python3 /app/python/pydnsperf_main.py --servers servers.txt --domains domains.txt --output /app/output/dnsperfpycxx --resolve-impl cxx
    # sleep 1
done

# sync data to SQL db
time /app/golang/dbsync --ldbpath /app/output/dnsperfgo --sqldbpath /app/output/dnsgo.db --verbose true 
time /app/golang/dbsync --ldbpath /app/output/dnsperfcxx --sqldbpath /app/output/dnscxx.db --verbose true 
time /app/golang/dbsync --ldbpath /app/output/dnsperfpy --sqldbpath /app/output/dnspy.db --verbose true 
# time /app/golang/dbsync --ldbpath /app/output/dnsperfpycxx --sqldbpath /app/output/dnspycxx.db --verbose true 


