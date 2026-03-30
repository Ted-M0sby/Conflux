package router

import "testing"

func TestTableMatch_PriorityAndPrefix(t *testing.T) {
	routes := []Route{
		{ID: "api", PathPrefix: "/user", Priority: 10, Targets: []string{"http://a.example"}},
		{ID: "vip", PathPrefix: "/user/vip", Priority: 100, Targets: []string{"http://b.example"}},
	}
	tab, err := NewTable(routes)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := tab.Match("/user/vip/x")
	if !ok || r.ID != "vip" {
		t.Fatalf("expected vip route, got %#v ok=%v", r, ok)
	}
	r2, ok := tab.Match("/user/other")
	if !ok || r2.ID != "api" {
		t.Fatalf("expected api route, got %#v ok=%v", r2, ok)
	}
}

func TestTableMatch_RootAndStripNormalization(t *testing.T) {
	routes := []Route{
		{ID: "x", PathPrefix: "api", Priority: 1, Targets: []string{"http://a.example"}},
	}
	tab, err := NewTable(routes)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := tab.Match("/api/v1")
	if !ok {
		t.Fatal("expected match")
	}
}
