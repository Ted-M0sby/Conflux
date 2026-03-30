package router

// Route defines a prefix-based proxy rule.
type Route struct {
	ID          string   `yaml:"id"`
	PathPrefix  string   `yaml:"path_prefix"`
	StripPrefix bool     `yaml:"strip_prefix"`
	Priority    int      `yaml:"priority"`
	Targets     []string `yaml:"targets"`
}

// RoutesFile is the root YAML document.
type RoutesFile struct {
	Routes []Route `yaml:"routes"`
}
