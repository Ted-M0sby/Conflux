#include "httplib.h"
#include <cstdlib>
#include <iostream>
#include <string>

static int env_port() {
  const char* v = std::getenv("MOCK_PORT");
  if (!v) return 18081;
  try { return std::stoi(v); } catch (...) { return 18081; }
}

int main() {
  int port = env_port();
  httplib::Server s;

  s.Get("/health", [](const httplib::Request&, httplib::Response& res) {
    res.set_content("ok", "text/plain");
  });

  s.Get(R"(/user/.*)", [](const httplib::Request& req, httplib::Response& res) {
    res.set_content(std::string("mock user path: ") + req.path, "text/plain");
  });

  s.Get(R"(.*)", [](const httplib::Request& req, httplib::Response& res) {
    res.set_content(std::string("mock path: ") + req.path, "text/plain");
  });

  std::cout << "mock backend listening :" << port << "\n";
  s.listen("0.0.0.0", port);
}
