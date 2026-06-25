// Package hashline implements LINE#ID content addressing for heretic.
//
// The `Read` tool appends a short hash to each line of its output, and the
// `hashline_edit` tool requires the hash to match the current file content
// before applying an edit. Stale edits (where the file changed between read
// and edit) are rejected, preventing the LLM from corrupting code that was
// modified out-of-band.
//
// Design notes:
//   - The hash alphabet is a 16-character subset: ZPMQVRWSNKTXJBYH.
//     This matches oh-my-opencode's hashline-core so the design is recognizable.
//   - The hash combines the line content with the previous and next lines to
//     reduce collisions on lines that appear identically in multiple places.
//   - We do NOT depend on oh-my-opencode's TypeScript source; the algorithm
//     is reimplemented here in Go using only stdlib (hash/fnv + character map).
package hashline

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// Alphabet is the 16-character hash alphabet used for LINE#ID.
const Alphabet = "ZPMQVRWSNKTXJBYH"

// hashLen is the number of characters in the rendered hash ID.
const hashLen = 4

// HashLine returns a 4-character hash of the given line content, mixed with
// the previous and next lines to reduce collisions on repeating content.
// Empty prev/next are allowed.
func HashLine(line, prev, next string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(prev))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(line))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(next))
	sum := h.Sum64()

	out := make([]byte, hashLen)
	for i := 0; i < hashLen; i++ {
		out[i] = Alphabet[sum&0xF]
		sum >>= 4
	}
	return string(out)
}

// TagLine returns "line#ID" where ID is the hash for the given line.
func TagLine(line, prev, next string) string {
	return line + "#" + HashLine(line, prev, next)
}

// TagReadOutput appends "#ID" hashes to each line of the given content.
// The hash for line N uses the actual previous and next lines from the input.
func TagReadOutput(content string) string {
	if content == "" {
		return ""
	}
	lines := strings.Split(content, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		var prev, next string
		if i > 0 {
			prev = lines[i-1]
		}
		if i < len(lines)-1 {
			next = lines[i+1]
		}
		out[i] = TagLine(line, prev, next)
	}
	return strings.Join(out, "\n")
}

// ErrHashMismatch is returned by ValidateEdit when the hash on the targeted
// line does not match the current file content.
type ErrHashMismatch struct {
	LineNumber int
	Line       string
	Expected   string
	Actual     string
}

func (e *ErrHashMismatch) Error() string {
	return fmt.Sprintf("hashline: hash mismatch on line %d: got %q want %q (line=%q)",
		e.LineNumber, e.Actual, e.Expected, e.Line)
}

// ExtractHash returns the trailing #XXXX hash on the line, or "" if none.
// Exported so the hashline_edit tool can use it.
func ExtractHash(line string) string {
	idx := strings.LastIndex(line, "#")
	if idx < 0 || idx == len(line)-1 {
		return ""
	}
	tail := line[idx+1:]
	if len(tail) != hashLen {
		return ""
	}
	for _, c := range tail {
		if !strings.ContainsRune(Alphabet, c) {
			return ""
		}
	}
	return tail
}

// ValidateLineHash checks that the line+hash pair the caller provides still
// matches the current file content. The caller passes the UNTAGGED line
// (the original text, without the #ID suffix) and the hash it expects.
// It returns nil on match, or an *ErrHashMismatch describing the discrepancy.
func ValidateLineHash(currentContent, targetLine, providedHash string) error {
	if providedHash == "" {
		return fmt.Errorf("hashline: provided hash is empty")
	}
	lines := strings.Split(currentContent, "\n")
	for i, line := range lines {
		if line != targetLine {
			continue
		}
		// Recompute the hash for this line and compare.
		var prev, next string
		if i > 0 {
			prev = lines[i-1]
		}
		if i < len(lines)-1 {
			next = lines[i+1]
		}
		gotHash := HashLine(line, prev, next)
		if gotHash != providedHash {
			return &ErrHashMismatch{LineNumber: i + 1, Line: line, Expected: providedHash, Actual: gotHash}
		}
		return nil
	}
	return fmt.Errorf("hashline: line not found in current content: %q", targetLine)
}

// ValidateStaleHash checks that a tagged line (line#ID) was tagged against
// the SAME content currently on disk. The caller passes the UNTAGGED line
// and the hash it was tagged with. The function recomputes the hash from
// the current content and returns ErrHashMismatch on mismatch.
func ValidateStaleHash(currentContent, targetLine, providedHash string) error {
	return ValidateLineHash(currentContent, targetLine, providedHash)
}
