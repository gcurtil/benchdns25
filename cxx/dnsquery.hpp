#include <stdio.h>
#include <string.h>
//#include <errno.h>
#include <arpa/inet.h>

#include <unbound.h>

#include <string>
#include <unordered_map>
#include <vector>

struct DnsServer 
{
public:
    DnsServer(std::string const& addr, std::string const& desc) : addr(addr), desc(desc) {}
    std::string addr;
    std::string desc;
};

struct DnsLookupResult 
{
    int ret;
    std::string ip;
    double lookup_time;
};

DnsLookupResult resolve_host_01(std::string const &dns_server, std::string const &host_to_resolve);

struct DnsResolverWithReuse
{
    DnsResolverWithReuse() 
    {
        // printf("DnsResolverWithReuse ctor, creating unbound context\n");        
    }
    ~DnsResolverWithReuse() 
    {
        for(auto it : ctx_map)
        {
            auto ctx = it.second;
            if (ctx)        
            {
                // printf("DnsResolverWithReuse dtor, deleting unbound context for <%s>\n", it.first.c_str());
                ub_ctx_delete(ctx);
            }
        }
    }
    
    DnsLookupResult resolve_host(std::string const& dns_server, std::string const& host_to_resolve);
    struct ub_ctx *get_ctx(std::string const &dns_server);

    // struct ub_ctx *ctx;
    std::unordered_map<std::string, struct ub_ctx*> ctx_map;
};

