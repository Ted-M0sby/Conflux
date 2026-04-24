#include "config.hpp"
#include <cstdlib>
#include <string>

static std::string env_or(const char* k, const std::string& d) {
  const char* v = std::getenv(k);
  return v ? std::string(v) : d;
}

static int env_int_or(const char* k, int d) {
  const char* v = std::getenv(k);
  if (!v) return d;
  try {
    return std::stoi(v);
  } catch (...) {
    return d;
  }
}

Config load_config() {
  Config c;
  c.routes_file = env_or("NEXUS_ROUTES_FILE", c.routes_file);
  c.lb = env_or("NEXUS_LB", c.lb);
  c.admin_prefix = env_or("NEXUS_ADMIN_PREFIX", c.admin_prefix);
  c.admin_token = env_or("NEXUS_ADMIN_TOKEN", "");
  c.rate_limit_enable = env_or("NEXUS_RATELIMIT_ENABLE", "true") != "false";
  c.rate_limit_rps = env_int_or("NEXUS_RATELIMIT_RPS", c.rate_limit_rps);
  c.rate_limit_window_sec = env_int_or("NEXUS_RATELIMIT_WINDOW_SEC", c.rate_limit_window_sec);
  return c;
}
