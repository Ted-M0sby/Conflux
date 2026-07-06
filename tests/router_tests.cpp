#include "router.hpp"

#include <cassert>
#include <fstream>

int main() {
  const char* file = "router_test_routes.yaml";
  std::ofstream out(file);
  out << "routes:\n"
      << "  - id: root\n"
      << "    path_prefix: /\n"
      << "    priority: 0\n"
      << "    targets:\n"
      << "      - http://127.0.0.1:18081\n"
      << "  - id: api\n"
      << "    path_prefix: /api\n"
      << "    strip_prefix: true\n"
      << "    priority: 10\n"
      << "    targets:\n"
      << "      - http://127.0.0.1:18082\n";
  out.close();

  RouterTable table;
  assert(table.load_yaml(file));
  assert(table.match("/api")->id == "api");
  assert(table.match("/api/users")->id == "api");
  assert(table.match("/apiary") == nullptr);

  return 0;
}
