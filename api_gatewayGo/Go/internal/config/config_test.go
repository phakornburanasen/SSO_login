package config

import "testing"

func TestParseLegacyServiceMap(t *testing.T) {
	snapshot, err := Parse([]byte(`{
		"O365": "http://127.0.0.1:5000",
		"Reports": "https://reports.example.com/base"
	}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got := snapshot.Names(); len(got) != 2 || got[0] != "O365" || got[1] != "Reports" {
		t.Fatalf("Names() = %#v", got)
	}

	service, ok := snapshot.Get("Reports")
	if !ok {
		t.Fatal("Reports service not found")
	}
	if service.URL != "https://reports.example.com/base" {
		t.Fatalf("Reports URL = %q", service.URL)
	}
}

func TestParseExtendedServiceMap(t *testing.T) {
	snapshot, err := Parse([]byte(`{
		"services": {
			"active": {"url": "http://active.example.com", "enabled": true},
			"disabled": {"url": "http://disabled.example.com", "enabled": false}
		}
	}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if _, ok := snapshot.Get("active"); !ok {
		t.Fatal("active service not found")
	}
	if _, ok := snapshot.Get("disabled"); ok {
		t.Fatal("disabled service should be hidden")
	}
	if got := snapshot.Names(); len(got) != 1 || got[0] != "active" {
		t.Fatalf("Names() = %#v", got)
	}
}
