#include "utils.hpp"
#include "date.h"

template <class T>
std::string join(std::vector<T> const &v, std::string const &sep)
{
    std::ostringstream os("");
    for (auto x : v)
    {
        if (!os.str().empty())
        {
            os << sep;
        }
        os << x;
    }
    return os.str();
}

std::string current_time_and_date_str2()
{    
    auto now = std::chrono::system_clock::now();
    // auto today = date::floor<date::days>(now);
    std::stringstream ss;
    // ss << today << ' ' << date::make_time(now - today) << " UTC";
    //auto today = date::year_month_day{date::floor<date::days>(now)};
    // ss << today << ' ' << date::format("%Y-%m-%d %T", date::floor<std::chrono::milliseconds>(now));    
    ss << date::format("%Y-%m-%d %T", date::floor<std::chrono::milliseconds>(now));    
    return ss.str();
    // date::year_month_day{ today } << ' ' << date::make_time(now - today)
    // auto res = date::format("%Y-%m-%d %T", date::floor<std::chrono::milliseconds>(now));
    // return res;
}

#include <ctime>
#include <iomanip>
std::string current_time_and_date_str()
{
    auto now = std::chrono::system_clock::now();
    auto in_time_t = std::chrono::system_clock::to_time_t(now);

    std::stringstream ss;
    ss << std::put_time(std::localtime(&in_time_t), "%Y-%m-%d %H:%M:%S");
    return ss.str();
}
