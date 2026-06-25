package hashline

import (
	"regexp"
	"strings"
)

// hashPattern matches the trailing #XXXX (4 chars from the alphabet).
var hashPattern = regexp.MustCompile(`#([ZPMQVRWSNKTXJBYH]{4})$`)

// StripHashes removes the trailing #ID hashes from tagged read output.
// Used to recover the original text from a tagged line.
func StripHashes(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = hashPattern.ReplaceAllString(l, "")
	}
	return strings.Join(out, "\n")
}

// DiffHunk is one segment of a diff.
type DiffHunk struct {
	Kind    DiffKind
	OldLine int    // 1-based; 0 for added lines
	NewLine int    // 1-based; 0 for removed lines
	Text    string // raw text (without leading +/-)
}

// DiffKind is the change type.
type DiffKind int

const (
	DiffContext DiffKind = iota
	DiffAdded
	DiffRemoved
)

// ComputeDiff produces a simple line-based diff. It is not a full
// Myers diff — it just identifies added/removed lines via set
// comparison. Good enough for inline display with the LLM.
func ComputeDiff(oldContent, newContent string) []DiffHunk {
	oldLines := splitLinesKeepContent(oldContent)
	newLines := splitLinesKeepContent(newContent)
	oldSet := make(map[string]int, len(oldLines))
	for i, l := range oldLines {
		oldSet[l] = i + 1
	}
	newSet := make(map[string]int, len(newLines))
	for i, l := range newLines {
		newSet[l] = i + 1
	}
	var hunks []DiffHunk
	// Removed: in old but not in new.
	for i, l := range oldLines {
		if _, ok := newSet[l]; !ok {
			hunks = append(hunks, DiffHunk{Kind: DiffRemoved, OldLine: i + 1, Text: l})
		}
	}
	// Added: in new but not in old.
	for i, l := range newLines {
		if _, ok := oldSet[l]; !ok {
			hunks = append(hunks, DiffHunk{Kind: DiffAdded, NewLine: i + 1, Text: l})
		}
	}
	return hunks
}

// FormatDiff renders a diff as a human-readable string with +/- prefixes.
func FormatDiff(hunks []DiffHunk) string {
	if len(hunks) == 0 {
		return "(no changes)"
	}
	var sb strings.Builder
	for _, h := range hunks {
		switch h.Kind {
		case DiffAdded:
			sb.WriteString("+ ")
		case DiffRemoved:
			sb.WriteString("- ")
		default:
			sb.WriteString("  ")
		}
		sb.WriteString(h.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatDiffWithHashes renders a diff with the new line's hashline hash
// appended to each added line. Useful as a richer response when an
// edit succeeds.
func FormatDiffWithHashes(hunks []DiffHunk) string {
	if len(hunks) == 0 {
		return "(no changes)"
	}
	var sb strings.Builder
	for i, h := range hunks {
		var prev, next string
		if i > 0 {
			prev = hunks[i-1].Text
		}
		if i < len(hunks)-1 {
			next = hunks[i+1].Text
		}
		switch h.Kind {
		case DiffAdded:
			sb.WriteString("+ ")
			sb.WriteString(h.Text)
			sb.WriteString("#")
			sb.WriteString(HashLine(h.Text, prev, next))
			sb.WriteString("\n")
		case DiffRemoved:
			sb.WriteString("- ")
			sb.WriteString(h.Text)
			sb.WriteString("\n")
		default:
			sb.WriteString("  ")
			sb.WriteString(h.Text)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// splitLinesKeepContent splits content on \n, keeping empty lines.
func splitLinesKeepContent(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
