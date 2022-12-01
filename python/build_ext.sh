# build wrapper module with pybind11
g++ -O3 -Wall -shared -std=c++14 -fPIC $(/usr/bin/python3 -m pybind11 --includes) \
  -o cxxdnsquery$(/usr/bin/python3-config --extension-suffix) \
  cxxdnsquery.cpp ../cxx/dnsquery.cpp ../cxx/utils.cpp -lunbound 

  
  

