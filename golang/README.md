# DNS Perf implementation in Go

## Commands

### Build


#### dnsperf

```shell
(benchdns) gui@monolith:~/devel/benchdnsdemo/golang$ go build cmd/dnsperf/dnsperf.go 
```

#### dbsync

```shell
(benchdns) gui@monolith:~/devel/benchdnsdemo/golang$ go build cmd/dbsync/dbsync.go 
```

### dnsperf - Run the DNS perf tests

Run the query

```shell
(benchdns) gui@monolith:~/devel/benchdnsdemo/golang$ time ./dnsperf --servers ../servers.txt --domains ../domains.txt --verbose=false 

servers: ../servers.txt, domains: ../domains.txt, output: dnsperfdb
run_benchmark, servers: ../servers.txt, domains: ../domains.txt, output: dnsperfdb
server_processor, s: <1.1.1.1, cloudflare1>, found parts: [1.1.1.1, cloudflare1 1.1.1.1 cloudflare1]
server_processor, s: <1.1.1.2, cloudflare2>, found parts: [1.1.1.2, cloudflare2 1.1.1.2 cloudflare2]
server_processor, s: <1.1.1.3, cloudflare3>, found parts: [1.1.1.3, cloudflare3 1.1.1.3 cloudflare3]
server_processor, s: <8.8.8.8, google1>, found parts: [8.8.8.8, google1 8.8.8.8 google1]
server_processor, s: <8.8.4.4, google2>, found parts: [8.8.4.4, google2 8.8.4.4 google2]
server_processor, s: <9.9.9.9, quad9>, found parts: [9.9.9.9, quad9 9.9.9.9 quad9]
server_processor, skipping line s: <# internal dns server below>
server_processor, s: <192.168.2.6, pf6>, found parts: [192.168.2.6, pf6 192.168.2.6 pf6]

real    0m1.995s
user    0m0.009s
sys     0m0.014s
```

### dbsync - sync from leveldb to sqlite

```shell
(benchdns) gui@monolith:~/devel/benchdnsdemo/golang$ time ./dbsync 
ldbpath: dnsperfdb
run_sync, ldbpath: dnsperfdb, sqldbpath: dns.db
run_sync, sqldb: &{0 {dns.db 0xc0000ba000} 0 {0 0} [] map[] 0 0 0xc00008e540 false map[] map[] 0 0 0 0 <nil> 0 0 0 0 0x4f8ea0}

real    0m0.089s
user    0m0.015s
sys     0m0.020s
```
