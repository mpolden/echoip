package geo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProvidersDefault(t *testing.T) {
	p, err := LoadProviders("")
	if err != nil {
		t.Fatalf("LoadProviders: %v", err)
	}
	if _, ok := p.Attributions["dbip"]; !ok {
		t.Error("expected dbip attribution in default config")
	}
	if _, ok := p.Attributions["maxmind"]; !ok {
		t.Error("expected maxmind attribution in default config")
	}
}

func TestResolveAutoDetect(t *testing.T) {
	p, err := LoadProviders("")
	if err != nil {
		t.Fatalf("LoadProviders: %v", err)
	}
	tests := []struct {
		name      string
		databases []Database
		want      []string
	}{
		{
			name:      "dbip country lite",
			databases: []Database{{Type: "DBIP-Country-Lite", Description: map[string]string{"en": "DB-IP.com - IP to Country"}}},
			want:      []string{"DB-IP"},
		},
		{
			name:      "legacy dbip type",
			databases: []Database{{Type: "IPV6-CITY"}},
			want:      []string{"DB-IP"},
		},
		{
			name:      "maxmind geolite2",
			databases: []Database{{Type: "GeoLite2-City"}},
			want:      []string{"MaxMind GeoLite2"},
		},
		{
			name:      "maxmind detected by description",
			databases: []Database{{Type: "SomeCustomType", Description: map[string]string{"en": "Created by MaxMind"}}},
			want:      []string{"MaxMind GeoLite2"},
		},
		{
			name:      "mixed providers are deduplicated in order",
			databases: []Database{{Type: "DBIP-City-Lite"}, {Type: "GeoLite2-Country"}, {Type: "DBIP-Country-Lite"}},
			want:      []string{"DB-IP", "MaxMind GeoLite2"},
		},
		{
			name:      "unknown type falls back to none",
			databases: []Database{{Type: "Unknown-DB"}},
			want:      []string{},
		},
		{
			name:      "no databases",
			databases: nil,
			want:      []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Resolve(tt.databases, "")
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d attributions %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i].Text != tt.want[i] {
					t.Errorf("attribution %d: got text %q, want %q", i, got[i].Text, tt.want[i])
				}
			}
		})
	}
}

func TestResolveForcedProvider(t *testing.T) {
	p, err := LoadProviders("")
	if err != nil {
		t.Fatalf("LoadProviders: %v", err)
	}
	got, err := p.Resolve([]Database{{Type: "GeoLite2-City"}}, "dbip")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 1 || got[0].Text != "DB-IP" {
		t.Errorf("forced dbip: got %v", got)
	}
	got, err = p.Resolve([]Database{{Type: "GeoLite2-City"}}, "none")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("forced none: got %v, want empty", got)
	}
	if _, err := p.Resolve(nil, "does-not-exist"); err == nil {
		t.Error("expected error for unknown forced provider")
	}
}

func TestLoadProvidersFromFile(t *testing.T) {
	config := `
attributions:
  example:
    text: "Example Geo"
    url: "https://example.com"
  none:
rules:
  - provider: example
    match:
      database_type: ["EXAMPLE-*"]
fallback: example
`
	file := filepath.Join(t.TempDir(), "providers.yaml")
	if err := os.WriteFile(file, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadProviders(file)
	if err != nil {
		t.Fatalf("LoadProviders: %v", err)
	}
	got, err := p.Resolve([]Database{{Type: "EXAMPLE-City"}}, "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 1 || got[0].Text != "Example Geo" || got[0].URL != "https://example.com" {
		t.Errorf("got %v", got)
	}
	// Unknown type hits fallback
	got, err = p.Resolve([]Database{{Type: "Whatever"}}, "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 1 || got[0].Text != "Example Geo" {
		t.Errorf("fallback: got %v", got)
	}
}

func TestLoadProvidersValidation(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(file, []byte("rules:\n  - provider: ghost\n    match:\n      database_type: ['*']\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProviders(file); err == nil {
		t.Error("expected error for rule referencing unknown provider")
	}
	if err := os.WriteFile(file, []byte("attributions:\n  x:\n    text: X\nfallback: ghost\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProviders(file); err == nil {
		t.Error("expected error for unknown fallback provider")
	}
	if _, err := LoadProviders(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("expected error for missing file")
	}
}
