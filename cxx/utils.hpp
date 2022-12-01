#include <chrono>
#include <memory>
#include <sstream>
#include <string>
#include <vector>

#include <cstdio>
#include <cstring>
#include <memory>
#include <string>
#include <stdexcept>


// https://github.com/HowardHinnant/date

// https://stackoverflow.com/questions/58864923/getting-current-time-with-millisecond-precision-using-put-time-in-c

// https://stackoverflow.com/questions/17223096/outputting-date-and-time-in-c-using-stdchrono
std::string current_time_and_date_str2();

std::string current_time_and_date_str();

template <class T>
std::string join(std::vector<T> const &v, std::string const &sep);

template<typename ... Args>
std::string string_format( const std::string& format, Args ... args )
{
    int size_s = std::snprintf( nullptr, 0, format.c_str(), args ... ) + 1; // Extra space for '\0'
    if( size_s <= 0 ){ throw std::runtime_error( "Error during formatting." ); }
    auto size = static_cast<size_t>( size_s );
    auto buf = std::make_unique<char[]>( size );
    std::snprintf( buf.get(), size, format.c_str(), args ... );
    return std::string( buf.get(), buf.get() + size - 1 ); // We don't want the '\0' inside
}


using TimePoint = std::chrono::system_clock::time_point;
using myclock = std::chrono::system_clock;
using myduration = std::chrono::duration<double>;

struct SimpleTimer
{
    SimpleTimer() : t_start(myclock::now()) {}
    
    double elapsed_seconds()
    {
        auto t_end = myclock::now();
        // std::chrono::duration<double> dt = t_end - t_start;
        myduration dt = t_end - t_start;
        return dt.count();
    }

    myclock::time_point t_start;
};
