// Package rules loads project-level rules from `.heretic/rules/*.md` and
// injects them into the system prompt.
//
// Design notes:
//   - One Rule per .md file in the rules directory.
//   - Missing directory returns an empty slice and nil error (no project
//     rules is a valid state).
//   - Non-.md files are skipped silently.
//   - We do NOT copy oh-my-opencode's rules-engine TypeScript code; this
//     is a small, Go-idiomatic reimplementation scoped to the feature port.
package rules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Rule is a single rule loaded from disk.
type Rule struct {
	// Name is the filename without the .md extension.
	Name string
	// Path is the absolute path to the rule file.
	Path string
	// Content is the raw file content (Markdown body, including any YAML
	// frontmatter).
	Content string
}

// LoadRules scans the rules directory for *.md files and returns them in
// sorted (by name) order. A missing directory is not an error: it returns
// (nil, nil). Any other error (e.g. permission denied) is returned.
func LoadRules(dir string) ([]Rule, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat rules dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("rules path %q is not a directory", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read rules dir %q: %w", dir, err)
	}
	var rules []Rule
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("read rule %q: %w", full, err)
		}
		rules = append(rules, Rule{
			Name:    strings.TrimSuffix(e.Name(), ".md"),
			Path:    full,
			Content: string(b),
		})
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Name < rules[j].Name
	})
	return rules, nil
}

// Inject concatenates the rule contents into the system prompt, each
// preceded by a header that names the rule. The order matches the order
// returned by LoadRules.
func Inject(systemPrompt string, rules []Rule) string {
	if len(rules) == 0 {
		return systemPrompt
	}
	var sb strings.Builder
	sb.WriteString(systemPrompt)
	for _, r := range rules {
		sb.WriteString("\n\n# Rule: ")
		sb.WriteString(r.Name)
		sb.WriteString("\n\n")
		sb.WriteString(r.Content)
	}
	return sb.String()
}
