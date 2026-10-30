package geo

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed providers.yaml
var defaultProvidersConfig []byte

// Attribution describes how a GeoIP data provider should be credited in the UI.
type Attribution struct {
	Text string `yaml:"text"`
	URL  string `yaml:"url"`
}

// Match is a set of glob patterns matched against MMDB metadata.
type Match struct {
	DatabaseType []string `yaml:"database_type"`
	Description  []string `yaml:"description"`
}

// Rule maps a match condition to a provider id.
type Rule struct {
	Provider string `yaml:"provider"`
	Match    Match  `yaml:"match"`
}

// Providers holds attribution metadata and database detection rules.
type Providers struct {
	Attributions map[string]*Attribution `yaml:"attributions"`
	Rules        []Rule                  `yaml:"rules"`
	Fallback     string                  `yaml:"fallback"`
}

// LoadProviders loads provider configuration from file. If file is empty,
// the built-in default configuration is used.
func LoadProviders(file string) (*Providers, error) {
	data := defaultProvidersConfig
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		data = b
	}
	var p Providers
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing geo providers config: %w", err)
	}
	if err := p.validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Providers) validate() error {
	for i, r := range p.Rules {
		if _, ok := p.Attributions[r.Provider]; !ok {
			return fmt.Errorf("rule %d references unknown provider %q", i, r.Provider)
		}
	}
	if p.Fallback != "" {
		if _, ok := p.Attributions[p.Fallback]; !ok {
			return fmt.Errorf("fallback references unknown provider %q", p.Fallback)
		}
	}
	return nil
}

// Resolve returns the deduplicated list of attributions for the given
// databases. If provider is not empty, it is used verbatim and no
// detection is performed.
func (p *Providers) Resolve(databases []Database, provider string) ([]Attribution, error) {
	ids := []string{}
	seen := map[string]struct{}{}
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if provider != "" {
		if _, ok := p.Attributions[provider]; !ok {
			return nil, fmt.Errorf("unknown geo provider %q", provider)
		}
		add(provider)
	} else {
		for _, db := range databases {
			id := p.match(db)
			if id == "" {
				id = p.Fallback
			}
			add(id)
		}
	}
	attributions := []Attribution{}
	for _, id := range ids {
		a := p.Attributions[id]
		if a == nil || a.Text == "" {
			// Provider without attribution (e.g. "none") renders nothing.
			continue
		}
		attributions = append(attributions, *a)
	}
	return attributions, nil
}

// match returns the id of the first rule matching db, or "" if none match.
func (p *Providers) match(db Database) string {
	for _, r := range p.Rules {
		if matchPatterns(r.Match.DatabaseType, db.Type) {
			return r.Provider
		}
		for _, description := range db.Description {
			if matchPatterns(r.Match.Description, description) {
				return r.Provider
			}
		}
	}
	return ""
}

// matchPatterns reports whether s matches any of the case-insensitive glob patterns.
func matchPatterns(patterns []string, s string) bool {
	s = strings.ToLower(s)
	for _, pattern := range patterns {
		if ok, err := path.Match(strings.ToLower(pattern), s); err == nil && ok {
			return true
		}
	}
	return false
}
