package router

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML parses routes from a YAML file.
func LoadYAML(path string) (*Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc RoutesFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	t, err := NewTable(doc.Routes)
	if err != nil {
		return nil, err
	}
	if len(t.Routes()) == 0 {
		return nil, fmt.Errorf("no valid routes in %s", path)
	}
	return t, nil
}

// ParseYAML parses routes from raw YAML bytes (e.g. Nacos payload).
func ParseYAML(data []byte) (*Table, error) {
	var doc RoutesFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return NewTable(doc.Routes)
}
