#pragma once
#include <string>

struct Config {
  std::string host = "0.0.0.0";
  int port = 8080;
  std::string routes_file = "configs/routes.yaml";
  std::string lb = "round_robin"; // round_robin|random|first
  std::string admin_prefix = "/admin";
  std::string admin_token = "";
  bool rate_limit_enable = true;
  int rate_limit_rps = 100;
  int rate_limit_window_sec = 10;
};

Config load_config();
