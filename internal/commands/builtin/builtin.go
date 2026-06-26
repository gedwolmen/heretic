// Package builtin embeds the OmO-style slash commands that ship with
// heretic. They are stored as markdown files under builtin/ and loaded
// at startup via embed.FS.
//
// The 9 builtin commands are:
//
//	ralph-loop      - self-referential dev loop
//	ulw-loop        - ultrawork loop with Oracle verification
//	cancel-ralph    - cancel an active Ralph Loop
//	refactor        - intelligent refactoring (LSP + AST-grep + tests)
//	start-work      - start Sisyphus work from a Prometheus plan
//	stop-continuation - stop all continuation mechanisms
//	remove-ai-slops - remove AI-generated code smells
//	handoff         - create a context handoff for a new session
//	hyperplan       - adversarial multi-agent planning
//
// Each command is a markdown file with YAML frontmatter:
//
//	---
//	description: (builtin) ...
//	argument-hint: '"task" [--flag=...]'
//	---
//	body...
package builtin

import (
	"embed"
	"io/fs"
	"regexp"
	"strings"

	"github.com/gedwolmen/heretic/internal/commands"
)

// namedArgPattern matches $ARG_NAME placeholders.
var namedArgPattern = regexp.MustCompile(`\$([A-Z][A-Z0-9_]*)`)

//go:embed *.md
var fsys embed.FS

// Load returns all builtin commands as CustomCommands. The Builtin
// field is set on each so the UI can group them separately.
func Load() ([]commands.CustomCommand, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var out []commands.CustomCommand
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		body, err := fsys.ReadFile(e.Name())
		if err != nil {
			return nil, err
		}
		// Strip the .md extension for the command name.
		name := strings.TrimSuffix(e.Name(), ".md")
		cmds := parse(name, string(body))
		cmds.Builtin = true
		cmds.ID = "builtin:" + name
		out = append(out, cmds)
	}
	return out, nil
}

// frontmatter delimiters
const (
	fmOpen  = "---"
	fmClose = "---"
)

// parse extracts description + argument-hint from frontmatter and
// returns a CustomCommand with the body. name is the file's basename
// (without .md) and is used as the command name.
func parse(name, content string) commands.CustomCommand {
	out := commands.CustomCommand{Name: name}
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != fmOpen {
		// No frontmatter — entire content is the body.
		out.Content = strings.TrimSpace(content)
		return out
	}
	// Find closing fence.
	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == fmClose {
			closeIdx = i
			break
		}
	}
	if closeIdx < 0 {
		// Unterminated frontmatter — treat as body.
		out.Content = strings.TrimSpace(content)
		return out
	}
	// Parse YAML-ish key: value lines.
	var desc, argHint string
	for _, line := range lines[1:closeIdx] {
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		// Strip a single pair of surrounding quotes so empty hints
		// like '' or "" parse to a genuinely empty string.
		val = strings.Trim(val, `"'`)
		switch key {
		case "description":
			desc = val
		case "argument-hint":
			argHint = val
		}
	}
	// Extract $ARG_NAME placeholders from the body to populate
	// Arguments (matching the existing custom-command convention).
	body := strings.TrimSpace(content[closeIdx+1:])
	out.Content = body
	out.ArgumentHint = argHint
	out.Arguments = extractArgNames(body)
	_ = desc
	return out
}

// extractArgNames returns the sorted unique $ARG_NAMES in content.
func extractArgNames(content string) []commands.Argument {
	matches := namedArgPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var args []commands.Argument
	for _, match := range matches {
		name := match[1]
		if !seen[name] {
			seen[name] = true
			args = append(args, commands.Argument{ID: name, Title: name, Required: true})
		}
	}
	return args
}
