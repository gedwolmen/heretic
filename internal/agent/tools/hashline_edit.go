package tools

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"

	"github.com/gedwolmen/heretic/internal/hashline"
)

//go:embed hashline_edit.md
var hashlineEditDescription string

// HashlineEditToolName is the LLM-facing name.
const HashlineEditToolName = "hashline_edit"

// HashlineEditParams mirrors edit params but requires a hash for the
// targeted line.
type HashlineEditParams struct {
	FilePath   string `json:"file_path" description:"The absolute path to the file to modify"`
	TargetLine string `json:"target_line" description:"The exact line content to find (without #ID)"`
	LineHash   string `json:"line_hash" description:"The #ID hash from the read output (4 chars)"`
	NewString  string `json:"new_string" description:"The replacement text"`
}

// NewHashlineEditTool returns a hashline-protected edit tool. The caller
// passes a `ReadFileFn` that returns the current on-disk content. This
// avoids coupling the tool to the coordinator's file IO.
func NewHashlineEditTool(readFile ReadFileFn) fantasy.AgentTool {
	if readFile == nil {
		readFile = defaultReadFile
	}
	return fantasy.NewAgentTool(
		HashlineEditToolName,
		hashlineEditDescription,
		func(ctx context.Context, params HashlineEditParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.FilePath == "" {
				return fantasy.NewTextErrorResponse("file_path is required"), nil
			}
			if params.TargetLine == "" {
				return fantasy.NewTextErrorResponse("target_line is required"), nil
			}
			if params.LineHash == "" {
				return fantasy.NewTextErrorResponse("line_hash is required (4 chars from #ID)"), nil
			}
			content, err := readFile(ctx, params.FilePath)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("read failed: %v", err)), nil
			}
			if err := hashline.ValidateLineHash(content, params.TargetLine, params.LineHash); err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			// Replace the target line (untagged form) with new_string.
			updated := strings.Replace(content, params.TargetLine, params.NewString, 1)
			if err := os.WriteFile(params.FilePath, []byte(updated), 0o644); err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("write failed: %v", err)), nil
			}
			return fantasy.NewTextResponse("hashline_edit applied"), nil
		},
	)
}

// ReadFileFn is the contract the tool uses to fetch current on-disk content.
type ReadFileFn func(ctx context.Context, path string) (string, error)

func defaultReadFile(ctx context.Context, path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
