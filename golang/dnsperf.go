package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/miekg/dns"
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

func run_benchmark(servers_path string, domains_path string, output_path string, verbose bool, numiter int) {
	fmt.Printf("run_benchmark, servers: %v, domains: %v, output: %v\n", servers_path, domains_path, output_path)

	// Read the list of DNS servers to use
	var servers []DnsServer
	r_servers := regexp.MustCompile(`^([0-9.]+),\s*(.*)\s*$`)
	r_comment := regexp.MustCompile(`^\s*#(.*)$`)

	var server_fn = func(s string) {
		//fmt.Printf("server_processor, got s: <%s>\n", s)
		if r_comment.MatchString(s) {
			fmt.Printf("server_processor, skipping line s: <%s>\n", s)
			return
		}
		parts := r_servers.FindStringSubmatch(s)
		if parts != nil {
			fmt.Printf("server_processor, s: <%s>, found parts: %v\n", s, parts)
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

	// query
	var counter int = 0
	for iter := 0; iter < numiter; iter++ {
		for _, ds := range servers {
			for _, domain := range domains {
				if verbose {
					fmt.Printf("DEBUG: calling resolve_host(%s, %s)\n", ds.addr, domain)
				}
				// auto now_str = current_time_and_date_str2();
				lres := resolve_host(ds.addr, domain)
				if verbose {
					fmt.Printf("DEBUG: resolve_host(%s, %s) returned %v\n",
						ds.addr, domain, lres)
				}

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

	fmt.Printf("servers: %v, domains: %v, output: %v\n", servers, domains, output)

	run_benchmark(servers, domains, output, verbose, numiter)
}
