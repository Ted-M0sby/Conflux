#include "router.hpp"
#include <algorithm>
#include <fstream>
#include <sstream>

static std::string trim(std::string s) {
  auto not_space = [](unsigned char c) { return !std::isspace(c); };
  s.erase(s.begin(), std::find_if(s.begin(), s.end(), not_space));
  s.erase(std::find_if(s.rbegin(), s.rend(), not_space).base(), s.end());
  return s;
}

static bool starts_with(const std::string& s, const std::string& p) {
  return s.size() >= p.size() && s.rfind(p, 0) == 0;
}

bool RouterTable::load_yaml(const std::string& file) {
  routes_.clear();

  std::ifstream in(file);
  if (!in.is_open()) return false;

  Route current;
  bool in_route = false;
  bool in_targets = false;
  std::string line;

  auto flush_current = [&]() {
    if (!in_route) return;
    if (current.path_prefix.empty()) current.path_prefix = "/";
    if (current.path_prefix[0] != '/') current.path_prefix = "/" + current.path_prefix;
    routes_.push_back(std::move(current));
    current = Route{};
    in_route = false;
    in_targets = false;
  };

  while (std::getline(in, line)) {
    auto t = trim(line);
    if (t.empty() || starts_with(t, "#")) continue;

    if (starts_with(t, "- id:")) {
      flush_current();
      in_route = true;
      current.id = trim(t.substr(5));
      continue;
    }

    if (!in_route) continue;

    if (starts_with(t, "path_prefix:")) {
      current.path_prefix = trim(t.substr(12));
      continue;
    }
    if (starts_with(t, "strip_prefix:")) {
      auto v = trim(t.substr(13));
      current.strip_prefix = (v == "true" || v == "True" || v == "1");
      continue;
    }
    if (starts_with(t, "priority:")) {
      try {
        current.priority = std::stoi(trim(t.substr(9)));
      } catch (...) {
        current.priority = 0;
      }
      continue;
    }
    if (starts_with(t, "targets:")) {
      in_targets = true;
      continue;
    }
    if (in_targets && starts_with(t, "- ")) {
      current.targets.push_back(trim(t.substr(2)));
      continue;
    }
    if (in_targets && !starts_with(t, "- ")) {
      in_targets = false;
    }
  }

  flush_current();

  std::sort(routes_.begin(), routes_.end(), [](const Route& a, const Route& b) {
    if (a.priority != b.priority) return a.priority > b.priority;
    return a.path_prefix.size() > b.path_prefix.size();
  });
  return true;
}

const Route* RouterTable::match(const std::string& path) const {
  for (const auto& r : routes_) {
    if (path == r.path_prefix) return &r;
    if (path.size() > r.path_prefix.size() &&
        path.rfind(r.path_prefix, 0) == 0 &&
        path[r.path_prefix.size()] == '/') {
      return &r;
    }
  }
  return nullptr;
}
