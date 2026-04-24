#pragma once
#include <string>
#include <vector>

struct Route {
  std::string id;
  std::string path_prefix;
  bool strip_prefix = false;
  int priority = 0;
  std::vector<std::string> targets;
};

class RouterTable {
public:
  bool load_yaml(const std::string& file);
  const Route* match(const std::string& path) const;
  const std::vector<Route>& routes() const { return routes_; }

private:
  std::vector<Route> routes_;
};
