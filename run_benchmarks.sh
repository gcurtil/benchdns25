#!/usr/bin/env bash

cd /app

# Run DNS perf queries
# for i in {1..5}; do 
n=5
for ((i=0; i < n; i++)); do
    echo "golang loop ${i} out of ${n}"
    time /app/golang/dnsperf --servers servers.txt --domains domains.txt --output /app/output/dnsperfgo --verbose=false
    sleep 1
done

# sync data to SQL db
time /app/golang/dbsync --ldbpath /app/output/dnsperfgo --sqldbpath /app/output/dnsgo.db --verbose true 







