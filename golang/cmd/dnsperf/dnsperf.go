package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/miekg/dns"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func read_lines(path string) []string {
	var lines []string
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		lines = append(lines, s)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

type line_processor func(s string)

func process_input_file(path string, fn line_processor) {
	lines := read_lines(path)

	for _, line := range lines {
		fn(line)
	}
}

type DnsServer struct {
	addr string
	desc string
}

type DnsLookupResult struct {
	ret         int
	ip          string
	lookup_time float64
}

func getAnswer(reply *dns.Msg) net.IP {
	for _, record := range reply.Answer {
		if record.Header().Rrtype == dns.TypeA {
			return record.(*dns.A).A
		}
	}
	return nil
}

func resolve_host(server_addr string, domain string) DnsLookupResult {
	start := time.Now()

	ret_val := 0
	ip_val := ""
	c := new(dns.Client)
	target := fmt.Sprintf("%s:53", server_addr)
	//in, rtt, err := c.Exchange(m1, "127.0.0.1:53")
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	reply, _, err := c.Exchange(msg, target)
	// fmt.Printf("resolve_host, target: %s, msg: %v\n", target, msg)
	// fmt.Printf("resolve_host, target: %s, c.Exchange returned err: %v, reply: %v\n", target, err, reply)

	if err != nil {
		// no answer
		ret_val = -1
	} else {
		if ip := getAnswer(reply); ip != nil {
			ip_val = ip.String()
		}
	}
	elapsed := time.Since(start)

	return DnsLookupResult{
		ret:         ret_val,
		ip:          ip_val,
		lookup_time: elapsed.Seconds(),
	}
}

func current_time_and_date_str2() string {
	s := time.Now().Format("2006-01-02 15:04:05.000")
	return s
}

func get_uuid() string {
	v, _ := uuid.NewRandom()
	return v.String()
}

func ldb_key(group string, id string) string {
	return fmt.Sprintf("%s|%s", group, id)
}

type ServerRecord struct {
	Addr string `json:"addr"`
	Desc string `json:"desc"`
}
type OutputRecord struct {
	// {
	// 	"server"        : { "addr" : server_addr, "desc" : server_desc },
	// 	"at"            : now_str,
	// 	"rid"           : ruuid_str,
	// 	"counter"       : counter,
	// 	"id"            : uuid_str,
	// 	"domain"        : domain,
	// 	"lookup_time"   : lres.lookup_time,
	// 	"lookup_ip"     : lres.ip,
	// }
	Server     ServerRecord `json:"server"`
	At         string       `json:"at"`
	Rid        string       `json:"rid"`
	Counter    int          `json:"counter"`
	Id         string       `json:"id"`
	Domain     string       `json:"domain"`
	LookupTime float64      `json:"lookup_time"`
	LookupIp   string       `json:"lookup_ip"`
}

func run_benchmark(servers_path string, domains_path string, output_path string, verbose bool, numiter int) {
	fmt.Printf("run_benchmark, servers: %v, domains: %v, output: %v, numiter: %d\n",
		servers_path, domains_path, output_path, numiter)

	// Read the list of DNS servers to use
	var servers []DnsServer
	r_servers := regexp.MustCompile(`^([0-9.]+),\s*(.*)\s*$`)
	r_comment := regexp.MustCompile(`^\s*#(.*)$`)

	var server_fn = func(s string) {
		//fmt.Printf("server_processor, got s: <%s>\n", s)
		if r_comment.MatchString(s) {
			// fmt.Printf("server_processor, skipping line s: <%s>\n", s)
			return
		}
		parts := r_servers.FindStringSubmatch(s)
		if parts != nil {
			// fmt.Printf("server_processor, s: <%s>, found parts: %v\n", s, parts)
			ds := DnsServer{addr: parts[1], desc: parts[2]}
			servers = append(servers, ds)
		}
	}
	process_input_file(servers_path, server_fn)

	// Read the list of domains to query
	var domains []string
	var domain_fn = func(s string) {
		domains = append(domains, s)
	}
	process_input_file(domains_path, domain_fn)

	// TODO warm up for lookups

	// Set up database connection information and open database
	o := &opt.Options{
		// Filter: filter.NewBloomFilter(10),
		Compression: opt.NoCompression,
	}
	db, err := leveldb.OpenFile(output_path, o)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	wo := opt.WriteOptions{}

	// Run the queries
	// take note of current time
	start_time_str := current_time_and_date_str2()
	_ = start_time_str
	ruuid_str := get_uuid()

	var counter int = 0
	for iter := 0; iter < numiter; iter++ {
		for _, ds := range servers {
			for _, domain := range domains {
				if verbose {
					fmt.Printf("DEBUG: calling resolve_host(%s, %s)\n", ds.addr, domain)
				}

				lres := resolve_host(ds.addr, domain)
				if verbose {
					fmt.Printf("DEBUG: resolve_host(%s, %s) returned %v\n",
						ds.addr, domain, lres)
				}

				var now_str string = current_time_and_date_str2()
				var uuid_str string = get_uuid()

				prettyJSON, json_err := json.MarshalIndent(OutputRecord{
					Server:     ServerRecord{Addr: ds.addr, Desc: ds.desc},
					At:         now_str,
					Rid:        ruuid_str,
					Counter:    counter,
					Id:         uuid_str,
					Domain:     domain,
					LookupTime: lres.lookup_time,
					LookupIp:   lres.ip,
				}, "", "\t")

				if verbose {
					fmt.Printf("DEBUG: resolve_host json_err: %v, output json: %v\n", json_err, string(prettyJSON))
				}
				dbKey := []byte(ldb_key(start_time_str, fmt.Sprintf("%012d", counter)))
				dbValue := prettyJSON
				db.Put(dbKey, dbValue, &wo)

				counter++
			}
		}
	}

}

func main() {
	// parse command line arguments
	var servers, domains, output string
	var numiter int
	var verbose bool
	flag.StringVar(&servers, "servers", "servers.txt", "File with DNS Servers to use")
	flag.StringVar(&domains, "domains", "domains.txt", "File with Domains to query")
	flag.StringVar(&output, "output", "dnsperfdb", "Output DB name")
	flag.IntVar(&numiter, "numiter", 1, "Number of Iterations")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.Parse()

	// fmt.Printf("servers: %v, domains: %v, output: %v\n", servers, domains, output)
	run_benchmark(servers, domains, output, verbose, numiter)
}
