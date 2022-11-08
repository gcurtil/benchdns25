#include <fstream>
#include <iostream>
#include <random>
#include <regex>
#include <string>
#include <vector>

#include "dnsquery.hpp"
#include "utils.hpp"

//////////////////////////
///// External deps  /////

// LevelDB for output
#include "leveldb/db.h"

// For cmd line argument parsing
#include "cxxopts.hpp"

// To convert to JSON
#include "json.hpp"
using json = nlohmann::json;


void usage(cxxopts::Options const& options, int ret_code)
{
    std::cout << options.help() << std::endl;
    exit(ret_code);
}

std::string get_uuid() {
    static std::random_device dev;
    static std::mt19937 rng(dev());

    std::uniform_int_distribution<int> dist(0, 15);

    const char *v = "0123456789abcdef";
    const bool dash[] = { 0, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0 };

    std::string res;
    res.reserve(20);
    for (int i = 0; i < 16; i++)
    {
        if (dash[i]) res += "-";
        res += v[dist(rng)];
        res += v[dist(rng)];
    }
    return res;
}

std::string ldb_key(std::string const& group, std::string const& id)
{
    return string_format("%s|%s", group.c_str(), id.c_str());
}

int 
run_benchmark
(
    std::string const& servers_path, 
    std::string const& domains_path,
    std::string const& output_path,
    bool verbose,
    int numiter
)
{
    // Read the list of DNS servers to use
    std::regex server_regex("^([0-9.]+),\\s*(.*)\\s*$");
    std::regex comment_regex("^\\s*#(.*)$");
    std::ifstream fis1(servers_path);
    std::string line;
    
    std::vector<DnsServer> servers;
    std::smatch m;
    while (std::getline(fis1, line))
    {        
        if (line.size() > 0)
        {
            if (std::regex_search(line, comment_regex)) 
            {
                std::cout << "run_benchmark, skipping line: <" << line << ">" << std::endl;
                continue;                
            }
            if (std::regex_search(line, m, server_regex)) 
            {
                // std::cout << m[1].str() << std::endl;
                // std::cout << m[2].str()  << std::endl;
                servers.emplace_back(m[1].str(), m[2].str());
            }
        }            
    }

    // Read the list of domains to query 
    std::vector<std::string> domains;
    std::ifstream fis2(domains_path);
    while (std::getline(fis2, line))
    {
        if (line.size() > 0) 
        {
            domains.push_back(line);
        }
    }

    auto ruuid = get_uuid();
    DnsResolverWithReuse resolverWithReuse;
    for(auto server : servers)
    {
        // warm up for lookups later
        auto _ = resolverWithReuse.get_ctx(server.addr);
    }

    auto start_time_str = current_time_and_date_str2();
    auto start_ldb_key = ldb_key(start_time_str, ruuid);
    std::cout << string_format("start_ldb_key: <%s>\n", start_ldb_key.c_str());

    // Set up database connection information and open database
    leveldb::DB* db;
    leveldb::Options options;
    options.create_if_missing = true;
    options.compression = leveldb::kNoCompression;

    leveldb::Status status = leveldb::DB::Open(options, output_path, &db);
    if (false == status.ok())
    {
        std::cerr << string_format("Unable to open/create output database <%s>\n", output_path.c_str());
        std::cerr << status.ToString() << std::endl;
        return -1;
    }
    leveldb::WriteOptions writeOptions;


    int counter = 0;
    for(auto iter = 0; iter < numiter; iter++)
    {
        for (auto server : servers)
        { 
            for(auto domain : domains)
            {
                if (verbose)
                {
                    std::cout << string_format("DEBUG: calling resolve_host(%s, %s)\n", 
                        server.addr.c_str(), domain.c_str());
                }
                auto now_str = current_time_and_date_str2();
                DnsLookupResult lres = resolverWithReuse.resolve_host(server.addr, domain);
                auto uuid = get_uuid();
                if (verbose)
                {
                    std::cout << string_format(
                        "NOW: %s, ID: %s, RID: %s, C:%d, SIP: %s, SD: %s, D: %s, DT: %.3f\n",
                        now_str.c_str(),
                        uuid.c_str(), ruuid.c_str(), counter,
                        server.addr.c_str(), server.desc.c_str(), domain.c_str(),
                        lres.lookup_time);
                }
                
                json j = {
                    {"server", {{"addr", server.addr}, {"desc", server.desc}}},
                    {"at", now_str},
                    {"rid", ruuid},
                    {"counter", counter},
                    {"id", uuid},
                    {"domain", domain},
                    {"lookup_time", lres.lookup_time},
                    {"lookup_ip", lres.ip},
                };

                auto dbKey = ldb_key(start_time_str, string_format("%012d", counter));
                auto dbValue = j.dump();
                if (verbose)
                {
                    std::cout << string_format("Key: <%s>, Value: %s\n", 
                                                dbKey.c_str(), dbValue.c_str());
                }
                db->Put(writeOptions, dbKey, dbValue);

                counter ++;
            }
        }
    }

    // Close the database
    delete db;
    
    return 0;
}

int main(int argc, char** argv)
{
    // https://github.com/jarro2783/cxxopts
    cxxopts::Options options("DnsPerf", "Measure DNS query performance for several servers");

    options.add_options()
        ("s,servers", "File with DNS Servers to use", cxxopts::value<std::string>()->default_value("servers.txt") )
        ("d,domains", "File with Domains to query", cxxopts::value<std::string>()->default_value("domains.txt"))
        ("o,output", "Output DB name", cxxopts::value<std::string>()->default_value("dnsperfdb"))
        ("n,numiter", "Number of Iterations", cxxopts::value<int>()->default_value("1"))
        ("v,verbose", "Verbose output", cxxopts::value<bool>()->default_value("false"))
        ("h,help", "Print usage")
    ;
    try 
    {
        auto result = options.parse(argc, argv);

        if (result.count("help"))
        {
            usage(options, 0);
        }

        std::string s1 = result["servers"].as<std::string>();
        std::string s2 = result["domains"].as<std::string>();
        std::string s3 = result["output"].as<std::string>();
        bool verbose = result["verbose"].as<bool>();
        int numiter = result["numiter"].as<int>();

        std::cout << string_format("Servers: <%s>, Domains: <%s>, Output: <%s>\n",
                                   s1.c_str(), s2.c_str(), s3.c_str())
                  << std::endl;

        run_benchmark(s1, s2, s3, verbose, numiter);
    }
    catch(cxxopts::OptionException const& e)
    {
        std::cout << "Error: " << e.what() << std::endl;
        usage(options, 1);
    }

}