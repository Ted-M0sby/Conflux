#include "httplib.h"
#include "config.hpp"
#include "router.hpp"

#include <chrono>
#include <deque>
#include <iomanip>
#include <iostream>
#include <mutex>
#include <random>
#include <sstream>
#include <string>
#include <unordered_map>

using namespace std::chrono;

static std::string json_escape(const std::string& s) {
  std::ostringstream o;
  for (char c : s) {
    switch (c) {
    case '"': o << "\\\""; break;
    case '\\': o << "\\\\"; break;
    case '\b': o << "\\b"; break;
    case '\f': o << "\\f"; break;
    case '\n': o << "\\n"; break;
    case '\r': o << "\\r"; break;
    case '\t': o << "\\t"; break;
    default:
      if (static_cast<unsigned char>(c) < 0x20) {
        o << "\\u" << std::hex << std::setw(4) << std::setfill('0') << (int)(unsigned char)c;
      } else {
        o << c;
      }
    }
  }
  return o.str();
}

class SlidingWindowLimiter {
public:
  SlidingWindowLimiter(int rps, int sec) : limit_(rps * sec), win_(seconds(sec)) {}

  bool allow(const std::string& key) {
    std::lock_guard<std::mutex> g(mu_);
    auto now = steady_clock::now();
    auto& q = m_[key];
    while (!q.empty() && now - q.front() > win_) q.pop_front();
    if ((int)q.size() >= limit_) return false;
    q.push_back(now);
    return true;
  }

private:
  int limit_;
  steady_clock::duration win_;
  std::mutex mu_;
  std::unordered_map<std::string, std::deque<steady_clock::time_point>> m_;
};

class Metrics {
public:
  void record(int status, long long latency_ms, bool limited) {
    std::lock_guard<std::mutex> g(mu_);
    ++total_;
    if (status >= 500) ++upstream_errors_;
    if (limited) ++rate_limited_;
    latency_total_ms_ += latency_ms;
  }

  std::string prometheus() const {
    std::lock_guard<std::mutex> g(mu_);
    std::ostringstream out;
    out << "# TYPE conflux_requests_total counter\n";
    out << "conflux_requests_total " << total_ << "\n";
    out << "# TYPE conflux_upstream_errors_total counter\n";
    out << "conflux_upstream_errors_total " << upstream_errors_ << "\n";
    out << "# TYPE conflux_rate_limited_total counter\n";
    out << "conflux_rate_limited_total " << rate_limited_ << "\n";
    out << "# TYPE conflux_request_latency_ms_avg gauge\n";
    out << "conflux_request_latency_ms_avg "
        << (total_ == 0 ? 0 : latency_total_ms_ / total_) << "\n";
    return out.str();
  }

private:
  mutable std::mutex mu_;
  long long total_ = 0;
  long long upstream_errors_ = 0;
  long long rate_limited_ = 0;
  long long latency_total_ms_ = 0;
};

struct TargetEndpoint {
  std::string host;
  int port = 80;
};

static bool parse_target(const std::string& target, TargetEndpoint& ep) {
  auto pos = target.find("://");
  auto hp = pos == std::string::npos ? target : target.substr(pos + 3);
  auto slash = hp.find('/');
  if (slash != std::string::npos) hp = hp.substr(0, slash);
  auto p2 = hp.rfind(':');
  ep.host = p2 == std::string::npos ? hp : hp.substr(0, p2);
  try {
    ep.port = p2 == std::string::npos ? 80 : std::stoi(hp.substr(p2 + 1));
  } catch (...) {
    return false;
  }
  return !ep.host.empty() && ep.port > 0;
}

static std::string query_string(const httplib::Request& req) {
  auto q = req.target.find('?');
  return q == std::string::npos ? "" : req.target.substr(q);
}

static std::string pick_target(const Route& r,
                               const std::string& lb,
                               std::unordered_map<std::string, size_t>& rr,
                               std::mt19937& gen) {
  if (r.targets.empty()) return "";
  if (lb == "first") return r.targets.front();
  if (lb == "random") {
    return r.targets[std::uniform_int_distribution<int>(0, (int)r.targets.size() - 1)(gen)];
  }
  auto& i = rr[r.id];
  auto v = r.targets[i % r.targets.size()];
  ++i;
  return v;
}

int main() {
  Config cfg = load_config();
  RouterTable table;
  if (!table.load_yaml(cfg.routes_file)) return 1;

  SlidingWindowLimiter limiter(cfg.rate_limit_rps, cfg.rate_limit_window_sec);
  Metrics metrics;
  std::unordered_map<std::string, size_t> rr;
  std::mutex lb_mu;
  std::mt19937 gen{std::random_device{}()};

  httplib::Server s;

  s.Get("/health", [](const httplib::Request&, httplib::Response& res) {
    res.set_content("ok", "text/plain");
  });

  s.Get((cfg.admin_prefix + "/routes").c_str(), [&](const httplib::Request& req, httplib::Response& res) {
    if (!cfg.admin_token.empty() && req.get_header_value("X-Admin-Token") != cfg.admin_token) {
      res.status = 401;
      res.set_content(R"({"error":"unauthorized"})", "application/json");
      return;
    }

    std::ostringstream out;
    out << "{\"routes\":[";
    bool first_route = true;
    for (const auto& r : table.routes()) {
      if (!first_route) out << ",";
      first_route = false;
      out << "{"
          << "\"id\":\"" << json_escape(r.id) << "\"," 
          << "\"path_prefix\":\"" << json_escape(r.path_prefix) << "\"," 
          << "\"strip_prefix\":" << (r.strip_prefix ? "true" : "false") << ","
          << "\"priority\":" << r.priority << ","
          << "\"targets\":[";
      bool first_target = true;
      for (const auto& t : r.targets) {
        if (!first_target) out << ",";
        first_target = false;
        out << "\"" << json_escape(t) << "\"";
      }
      out << "]}";
    }
    out << "]}";
    res.set_content(out.str(), "application/json");
  });

  s.Get("/metrics", [&](const httplib::Request&, httplib::Response& res) {
    res.set_content(metrics.prometheus(), "text/plain");
  });

  auto proxy = [&](const httplib::Request& req, httplib::Response& res) {
    auto start = steady_clock::now();
    bool limited = false;
    std::string route_id = "-";
    std::string target = "-";

    auto finish = [&]() {
      auto latency = duration_cast<milliseconds>(steady_clock::now() - start).count();
      metrics.record(res.status, latency, limited);
      std::cout << req.method << " " << req.target
                << " route=" << route_id
                << " target=" << target
                << " status=" << res.status
                << " latency_ms=" << latency << "\n";
    };

    if (cfg.rate_limit_enable && !limiter.allow(req.remote_addr)) {
      limited = true;
      res.status = 429;
      res.set_content(R"({"error":"rate_limit"})", "application/json");
      finish();
      return;
    }

    auto* r = table.match(req.path);
    if (!r) {
      res.status = 404;
      res.set_content("not found", "text/plain");
      finish();
      return;
    }
    route_id = r->id;

    {
      std::lock_guard<std::mutex> g(lb_mu);
      target = pick_target(*r, cfg.lb, rr, gen);
    }
    if (target.empty()) {
      res.status = 502;
      res.set_content("bad gateway", "text/plain");
      finish();
      return;
    }

    TargetEndpoint ep;
    if (!parse_target(target, ep)) {
      res.status = 502;
      res.set_content("bad gateway", "text/plain");
      finish();
      return;
    }

    std::string out_path = req.path;
    if (r->strip_prefix && out_path.rfind(r->path_prefix, 0) == 0) {
      out_path = out_path.substr(r->path_prefix.size());
      if (out_path.empty() || out_path[0] != '/') out_path = "/" + out_path;
    }
    out_path += query_string(req);

    httplib::Client cli(ep.host, ep.port);
    cli.set_connection_timeout(std::chrono::milliseconds(cfg.proxy_timeout_ms));
    cli.set_read_timeout(std::chrono::milliseconds(cfg.proxy_timeout_ms));
    httplib::Request up_req;
    up_req.method = req.method;
    up_req.path = out_path;
    up_req.headers = req.headers;
    up_req.headers.erase("Host");
    up_req.headers.erase("Connection");
    up_req.body = req.body;
    auto up = cli.send(up_req);
    if (!up) {
      res.status = 502;
      res.set_content("bad gateway", "text/plain");
      finish();
      return;
    }

    res.status = up->status;
    for (auto& h : up->headers) res.set_header(h.first.c_str(), h.second.c_str());
    res.body = up->body;
    finish();
  };

  s.Get(R"(.*)", proxy);
  s.Post(R"(.*)", proxy);
  s.Put(R"(.*)", proxy);
  s.Delete(R"(.*)", proxy);
  s.Patch(R"(.*)", proxy);

  std::cout << "listening :" << cfg.port << "\n";
  s.listen(cfg.host, cfg.port);
}
