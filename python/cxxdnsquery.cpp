#include <pybind11/pybind11.h>

namespace py = pybind11;

#include "../cxx/dnsquery.hpp"
// We want to expose the following function:
// DnsLookupResult resolve_host_01(std::string const &dns_server, std::string const &host_to_resolve)

int add(int i, int j) {
    return i + j;
}

PYBIND11_MODULE(cxxdnsquery, m) {
    m.doc() = "pybind11 DNS query"; // optional module docstring

    m.def("add", &add, "A function which adds two numbers",
      py::arg("i"), py::arg("j"));

    py::class_<DnsLookupResult>(m, "DnsLookupResult")
        .def(py::init<>())
        .def_readwrite("ret", &DnsLookupResult::ret)
        .def_readwrite("ip", &DnsLookupResult::ip)
        .def_readwrite("lookup_time", &DnsLookupResult::lookup_time)    
    ;

    // resolve_host_01(std::string const &dns_server, std::string const &host_to_resolve)
    m.def("resolve_host_cxx", &resolve_host_01, "resolve_host impl in c++",
          py::arg("dns_server"), py::arg("host_to_resolve"))
    ;
}
