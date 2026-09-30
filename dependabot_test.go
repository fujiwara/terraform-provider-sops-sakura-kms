package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

const catchAllGroup = "others"

type dependabotGroup struct {
	Patterns        []string `yaml:"patterns"`
	ExcludePatterns []string `yaml:"exclude-patterns"`
}

type namedGroup struct {
	name string
	dependabotGroup
}

// gomodGroups returns the groups of the gomod update in dependabot.yml in the defined order.
func gomodGroups(t *testing.T) []namedGroup {
	t.Helper()
	b, err := os.ReadFile(".github/dependabot.yml")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Updates []struct {
			PackageEcosystem string    `yaml:"package-ecosystem"`
			Groups           yaml.Node `yaml:"groups"`
		} `yaml:"updates"`
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	for _, u := range cfg.Updates {
		if u.PackageEcosystem != "gomod" {
			continue
		}
		var groups []namedGroup
		// Groups is a mapping node; Content holds key and value nodes alternately.
		for i := 0; i+1 < len(u.Groups.Content); i += 2 {
			g := namedGroup{name: u.Groups.Content[i].Value}
			if err := u.Groups.Content[i+1].Decode(&g.dependabotGroup); err != nil {
				t.Fatal(err)
			}
			groups = append(groups, g)
		}
		return groups
	}
	t.Fatal("gomod update is not found in dependabot.yml")
	return nil
}

// directDependencies returns the module paths of direct dependencies in go.mod.
func directDependencies(t *testing.T) []string {
	t.Helper()
	out, err := exec.Command("go", "mod", "edit", "-json").Output()
	if err != nil {
		t.Fatal(err)
	}
	var mod struct {
		Require []struct {
			Path     string
			Indirect bool
		}
	}
	if err := json.Unmarshal(out, &mod); err != nil {
		t.Fatal(err)
	}
	var deps []string
	for _, r := range mod.Require {
		if !r.Indirect {
			deps = append(deps, r.Path)
		}
	}
	return deps
}

// matchAny reports whether name matches any of Dependabot's patterns ("*" matches any characters).
func matchAny(patterns []string, name string) bool {
	for _, p := range patterns {
		parts := strings.Split(p, "*")
		for i := range parts {
			parts[i] = regexp.QuoteMeta(parts[i])
		}
		if regexp.MustCompile("^" + strings.Join(parts, ".*") + "$").MatchString(name) {
			return true
		}
	}
	return false
}

// TestDependabotDirectDependencies ensures that direct dependencies in go.mod
// are not grouped into the catch-all group, so that they are updated individually
// or in a dedicated group.
func TestDependabotDirectDependencies(t *testing.T) {
	groups := gomodGroups(t)
	deps := directDependencies(t)
	if len(deps) == 0 {
		t.Fatal("no direct dependencies are found in go.mod")
	}
	for _, dep := range deps {
		for _, g := range groups {
			if !matchAny(g.Patterns, dep) || matchAny(g.ExcludePatterns, dep) {
				continue
			}
			// Dependabot assigns a dependency to the first matching group.
			if g.name == catchAllGroup {
				t.Errorf("direct dependency %s is grouped into %q; add it to exclude-patterns in .github/dependabot.yml", dep, catchAllGroup)
			}
			break
		}
	}
}

// TestDependabotExcludePatterns ensures that exclude-patterns of the catch-all
// group do not contain stale entries.
func TestDependabotExcludePatterns(t *testing.T) {
	deps := directDependencies(t)
	for _, g := range gomodGroups(t) {
		if g.name != catchAllGroup {
			continue
		}
		for _, p := range g.ExcludePatterns {
			if !matchAnyDep(p, deps) {
				t.Errorf("exclude-pattern %q in .github/dependabot.yml does not match any direct dependency; remove it", p)
			}
		}
		return
	}
	t.Fatalf("group %q is not found in dependabot.yml", catchAllGroup)
}

func matchAnyDep(pattern string, deps []string) bool {
	for _, dep := range deps {
		if matchAny([]string{pattern}, dep) {
			return true
		}
	}
	return false
}
