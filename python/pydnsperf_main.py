import argparse
import datetime
import json
import logging
import pathlib
import re
import uuid
from dataclasses import dataclass
from typing import Any, Dict, Union

import dns.resolver
# https://plyvel.readthedocs.io/en/1.3.0/user.html
import plyvel
from util import Timer

# for C++ implementation 
import cxxdnsquery


logging.basicConfig(level=logging.DEBUG,
                    format='%(asctime)s %(levelname)s %(module)s %(message)s',
                    handlers=[logging.StreamHandler()])


def current_time_and_date_str2() -> str:
    '''Return the current UTC time as a string.
    '''
    t = datetime.datetime.utcnow()
    return t.strftime("%Y-%m-%d %T.%f")[:-3]

def get_uuid() -> str:
    '''Return a uuid as a string.    
    '''
    return uuid.uuid4().hex

def ldb_key(group:str, id_str: str) -> str:
    '''Create a composite key for level db using the 2 parameters passed in.
    '''
    return f"{group}|{id_str}"


@dataclass
class LookupResult:
    '''Holds the result of the DNS lookup.
    '''
    ret: int
    ip: str
    lookup_time: float


def resolve_host(server_addr: str, domain: str, resolver_cache: Union[Dict[str, Any], None] = None, verbose=None) -> LookupResult:
    '''Resolves the host name passed in using the DNS server specified, and optionally a cache for the Resolver objects.
    '''
    with Timer("resolve_host, server_addr: %s, domain: %s", server_addr, domain, verbose=verbose) as t1:
        # if cache passed in, check if we have a cached resolver for this server
        resolver = resolver_cache.get(server_addr) if resolver_cache is not None else None
        if not resolver:
            # logging.info("resolve_host, creating new resolver for server_addr: %s", server_addr)
            resolver = dns.resolver.Resolver(configure=False)
            resolver.nameservers = [ server_addr ]
        else:
            #logging.info("resolve_host, reusing resolver for server_addr: %s", server_addr)
            pass

        # if cache passed in, save the resolver
        if resolver_cache is not None:
            resolver_cache[server_addr] = resolver

        answer = resolver.resolve(domain, "A")

    ip_addr = answer[0].address if answer else ""
    if verbose:
        logging.info("resolve_host, server_addr: %s, domain: %s, answer: <%r>, ip_addr: %s",
                server_addr, domain, answer, ip_addr)
        
    res = LookupResult(ret=0, ip=ip_addr, lookup_time=t1.elapsed())
    return res


def resolve_host_with_cxx(server_addr: str, domain: str, resolver_cache: Union[Dict[str, Any], None] = None, verbose=None) -> LookupResult:
    '''Resolves the host name passed in using the DNS server specified. 
       Uses implementation in C++ (note the cache for the Resolver objects is not used in this case).
    '''
    _ = resolver_cache
    with Timer("resolve_host_cxx, server_addr: %s, domain: %s", server_addr, domain, verbose=verbose) as t1:
        # call resolve function through cxxdnsquery module
        r = cxxdnsquery.resolve_host_cxx(server_addr, domain)

    ip_addr = r.ip
    if verbose:
        logging.info("resolve_host_cxx, server_addr: %s, domain: %s, r: <%r>, ip_addr: %s",
                server_addr, domain, r, ip_addr)
    
    res = LookupResult(ret=0, ip=ip_addr, lookup_time=t1.elapsed())
    return res


def bench_dnsquery(servers_path: str, domains_path: str, output_path: str, verbose: bool, 
    numiter: int, resolve_impl: str):
    '''Run the DNS query benchmark for the list of servers and domains specified.
    '''
    # read list of servers
    re_server = re.compile("^([0-9.]+),\\s*(.*)\\s*$")
    re_comment = re.compile("^\\s*#(.*)$")
    server_data = []
    s = pathlib.Path(servers_path).read_text(encoding='utf-8')
    for line in s.splitlines():
        if re_comment.search(line):
            continue
        m = re_server.search(line)
        if m:
            server_data.append((m.group(1), m.group(2)))

    logging.info("server_data: %d servers, list: <%s>", len(server_data), server_data)

    # read list of domains
    domain_data = []
    s = pathlib.Path(domains_path).read_text(encoding='utf-8')
    for line in s.splitlines():
        if len(line) > 0:
            domain_data.append(line)

    logging.info("domain_data: %d domains, list: <%s>", len(domain_data), domain_data)

    resolver_cache = {}
    # TODO: warmup

    # take not of current time
    start_time_str: str = current_time_and_date_str2()

    db = plyvel.DB(output_path, create_if_missing=True)
    ruuid_str = get_uuid()
    counter = 0

    # check which implementation we want to use for resolving DNS queries, python or c++
    if resolve_impl == "python":
        resolve_host_fn = resolve_host
    elif resolve_impl == "cxx":
        resolve_host_fn = resolve_host_with_cxx
    else:
        raise Exception(f"Invalid value for resolve_impl: {resolve_impl}")

    with db.write_batch() as wb:
        for _ in range(numiter):
            for server_addr, server_desc in server_data:
                for domain in domain_data:
                    if verbose:
                        logging.info("DEBUG: calling resolve_host(%s, %s)\n", server_addr, domain)
                    lres = resolve_host_fn(server_addr, domain, resolver_cache=resolver_cache, verbose=verbose)
                    
                    now_str: str = current_time_and_date_str2()
                    uuid_str: str = get_uuid()
                    if verbose:
                        logging.info("DEBUG: NOW: %s, ID: %s, RID: %s, C:%d, SIP: %s, SD: %s, D: %s, DT: %.3f\n",
                            now_str,
                            uuid_str, ruuid_str, counter,
                            server_addr, server_desc, domain,
                            lres.lookup_time)

                    json_d = {
                        "server"        : { "addr" : server_addr, "desc" : server_desc },
                        "at"            : now_str,
                        "rid"           : ruuid_str,
                        "counter"       : counter,
                        "id"            : uuid_str,
                        "domain"        : domain,
                        "lookup_time"   : lres.lookup_time,
                        "lookup_ip"     : lres.ip,
                    }

                    dbKey = ldb_key(start_time_str, f"{counter:012d}")
                    dbValue = json.dumps(json_d)
                    if verbose:
                        logging.info("DEBUG Key: <%s>, Value: %s\n", dbKey, dbValue)
                    # writes as bytes
                    wb.put(dbKey.encode(), dbValue.encode())

                    counter += 1


    # Close the database
    db.close()


def main():
    parser = argparse.ArgumentParser(description='DNS query performance tester')
    parser.add_argument('--servers', "-s", default="servers.txt", help='File with DNS Servers to use')
    parser.add_argument('--domains', "-d", default="domains.txt", help='File with Domains to query')
    parser.add_argument('--output',  "-o", default="dnsperfdb_py", help='Output DB name')
    parser.add_argument("--verbose", "-v", default=False, action="store_true", help="Verbose output")
    parser.add_argument('--numiter', "-n", default=1, type=int, help='Number of Iterations')
    parser.add_argument('--resolve-impl', default="python", help='Implementation for DNS resolving, <python|cxx>')
    args = parser.parse_args()

    # df = bench_leveldb(args.redis_host, args.redis_port, args.num_runs)
    bench_dnsquery(args.servers, args.domains, args.output, args.verbose, args.numiter, args.resolve_impl)

if __name__ == '__main__':
    main()
