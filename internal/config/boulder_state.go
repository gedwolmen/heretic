package config

// BoulderStateConfig configures the boulder feature: a JSON-backed work
// tracker that survives across sessions. Mirrors
// packages/omo-opencode/src/config/schema/boulder-state.ts and
// packages/boulder-state/.
//
// The boulder state file is named `boulder.json` and lives at the project
// root (or in a configured location). It tracks a single active work item
// plus its history.
type BoulderStateConfig struct {
	// Enabled toggles the boulder feature. Default: false.
	Enabled bool `json:"enabled,omitempty" jsonschema:"description=Enable boulder state tracking,default=false"`

	// StateFile overrides the default `boulder.json` location.
	StateFile string `json:"state_file,omitempty" jsonschema:"description=Override the default boulder state file location"`

	// MaxHistory limits how many completed/regressed entries are kept
	// in the history slice. Default: 100.
	MaxHistory int `json:"max_history,omitempty" jsonschema:"description=Maximum history entries to retain,default=100,minimum=1"`

	// AutoAdvance on completion moves the boulder to the next pending
	// entry automatically. Default: true.
	AutoAdvance bool `json:"auto_advance,omitempty" jsonschema:"description=Auto-advance to next pending entry on completion,default=true"`

	// WorktreeEnabled creates a per-boulder git worktree so feature work
	// happens on a branch. Default: true.
	WorktreeEnabled bool `json:"worktree_enabled,omitempty" jsonschema:"description=Create a per-boulder git worktree,default=true"`

	// WorktreeBaseBranch is the base branch to fork from when creating
	// a worktree. Default: current HEAD.
	WorktreeBaseBranch string `json:"worktree_base_branch,omitempty" jsonschema:"description=Base branch for new boulder worktrees (default: current HEAD)"`
}

// EffectiveMaxHistory returns MaxHistory with default.
func (b BoulderStateConfig) EffectiveMaxHistory() int {
	if b.MaxHistory < 1 {
		return 100
	}
	return b.MaxHistory
}
