#include "dnsquery.hpp"
#include "utils.hpp"


DnsLookupResult resolve_host_01(std::string const &dns_server, std::string const &host_to_resolve)
{
    struct ub_ctx *ctx;
    struct ub_result *result;
    int retval;

    /* create context */
    ctx = ub_ctx_create();
    if (!ctx)
    {
        printf("error: could not create unbound context\n");
        return DnsLookupResult({1, "", 0});
    }

    retval = ub_ctx_set_fwd(ctx, dns_server.c_str());
    if (retval != 0)
    {
        printf("error: could not call ub_ctx_set_fwd \n");
        return DnsLookupResult({1, "", 0});
    }


    SimpleTimer timer1;    
    /* query for webserver */
    retval = ub_resolve(ctx, host_to_resolve.c_str(),
                        1 /* TYPE A (IPv4 address) */,
                        1 /* CLASS IN (internet) */, &result);
    if (retval != 0)
    {
        printf("resolve error: %s\n", ub_strerror(retval));
        return DnsLookupResult({1, "", 0});
    }

    std::string resolved_ip;
    /* show first result */
    if (result->havedata)
    {
        resolved_ip = inet_ntoa(*(struct in_addr *)result->data[0]);
        // printf("The address is %s\n", resolved_ip.c_str());
    }
    double dt1 = timer1.elapsed_seconds() * 1.0;

    ub_resolve_free(result);
    ub_ctx_delete(ctx);    

    return DnsLookupResult({0, resolved_ip, dt1});
}

struct ub_ctx* 
DnsResolverWithReuse::get_ctx(std::string const& dns_server) {
    //auto ctx = ctx_map[dns_server];        
    //if (!ctx) {
    ub_ctx* ctx = nullptr;
    auto it = ctx_map.find(dns_server);
    if (it == ctx_map.end())
    {
        
        SimpleTimer t1;
        // printf("DnsResolverWithReuse get_ctx, creating ctx for %s\n", dns_server.c_str());
        ctx = ub_ctx_create();
        //printf("DnsResolverWithReuse get_ctx, ub_ctx_create for %s: %.3f ms\n", dns_server.c_str(), t1.elapsed_seconds());
        ctx_map[dns_server] = ctx;
        SimpleTimer t2;
        int retval = ub_ctx_set_fwd(ctx, dns_server.c_str());
        //printf("DnsResolverWithReuse get_ctx, ub_ctx_set_fwd for %s: %.3f ms\n", dns_server.c_str(), t2.elapsed_seconds());
        if (retval != 0)
        {
            printf("error: could not call ub_ctx_set_fwd \n");
            ctx = nullptr;
            ctx_map[dns_server] = ctx;
        }
        
    }
    else 
    {
        ctx = it->second;
    }
    
    return ctx;
}

DnsLookupResult 
DnsResolverWithReuse::resolve_host(std::string const& dns_server, std::string const& host_to_resolve)
{    
    struct ub_result *result;
    int retval;

    auto ctx = get_ctx(dns_server);
    /* create context */    
    if (!ctx)
    {
        printf("error: could not get valid unbound context\n");
        return DnsLookupResult({1, "", 0});
    }

    // retval = ub_ctx_set_fwd(ctx, dns_server.c_str());
    // if (retval != 0)
    // {
    //     printf("error: could not call ub_ctx_set_fwd \n");
    //     return DnsLookupResult({1, "", 0});
    // }

    SimpleTimer timer1;    
    /* query for webserver */
    retval = ub_resolve(ctx, host_to_resolve.c_str(),
                        1 /* TYPE A (IPv4 address) */,
                        1 /* CLASS IN (internet) */, &result);
    if (retval != 0)
    {
        printf("resolve error: %s\n", ub_strerror(retval));
        return DnsLookupResult({1, "", 0});
    }

    std::string resolved_ip;
    /* show first result */
    if (result->havedata)
    {
        resolved_ip = inet_ntoa(*(struct in_addr *)result->data[0]);
        // printf("The address is %s\n", resolved_ip.c_str());
    }
    double dt1 = timer1.elapsed_seconds() * 1.0;

    ub_resolve_free(result);    

    return DnsLookupResult({0, resolved_ip, dt1});
}
