package team

import (
	"github.com/gedwolmen/heretic/internal/agent/builtin"
)

// Eligibility is the team member eligibility classification.
// Mirrors AGENT_ELIGIBILITY_REGISTRY from
// packages/omo-opencode/src/features/team-mode/types.ts.
type Eligibility string

const (
	// Eligible members may be added to a team.
	Eligible Eligibility = "eligible"
	// Conditional members may be added with extra setup.
	Conditional Eligibility = "conditional"
	// HardReject members are never allowed in a team.
	HardReject Eligibility = "hard-reject"
)

// EligibilityEntry is one (agent, eligibility) pair.
type EligibilityEntry struct {
	Agent       builtin.Name
	Eligibility Eligibility
	// Reason is a short human-readable note.
	Reason string
}

// DefaultEligibility returns the standard eligibility list.
//
//	eligible:     sisyphus, atlas, sisyphus-junior
//	conditional:  hephaestus
//	hard-reject:  oracle, librarian, explore, multimodal-looker, metis, momus, prometheus
func DefaultEligibility() []EligibilityEntry {
	return []EligibilityEntry{
		{Agent: builtin.NameSisyphus, Eligibility: Eligible},
		{Agent: builtin.NameHephaestus, Eligibility: Conditional, Reason: "lacks teammate: \"allow\" permission by default — apply D-36 in tool-config-handler.ts or use subagent_type: \"sisyphus\""},
		{Agent: builtin.NameOracle, Eligibility: HardReject, Reason: "use task/delegate-task"},
		{Agent: builtin.NameLibrarian, Eligibility: HardReject},
		{Agent: builtin.NameExplore, Eligibility: HardReject},
		{Agent: builtin.NameMultimodalLooker, Eligibility: HardReject},
		{Agent: builtin.NameMetis, Eligibility: HardReject, Reason: "pre-planning — not a team member"},
		{Agent: builtin.NameMomus, Eligibility: HardReject, Reason: "plan review — not a team member"},
		{Agent: builtin.NameAtlas, Eligibility: Eligible},
		{Agent: builtin.NameSisyphusJunior, Eligibility: Eligible},
		{Agent: builtin.NamePrometheus, Eligibility: HardReject},
	}
}

// IsEligible returns true if the given agent may be added to a team.
// Conditional members are not auto-eligible; the caller must opt in.
func IsEligible(agent builtin.Name) bool {
	for _, e := range DefaultEligibility() {
		if e.Agent == agent && e.Eligibility == Eligible {
			return true
		}
	}
	return false
}

// IsAllowed returns true if the agent is eligible OR conditional.
// Conditional callers must explicitly request inclusion.
func IsAllowed(agent builtin.Name) bool {
	for _, e := range DefaultEligibility() {
		if e.Agent == agent && e.Eligibility != HardReject {
			return true
		}
	}
	return false
}

// GetEntry returns the eligibility entry for an agent.
func GetEntry(agent builtin.Name) (EligibilityEntry, bool) {
	for _, e := range DefaultEligibility() {
		if e.Agent == agent {
			return e, true
		}
	}
	return EligibilityEntry{}, false
}

// Re-exports for callers that want string-typed agent names.
const (
	AgentSisyphus         = builtin.NameSisyphus
	AgentHephaestus       = builtin.NameHephaestus
	AgentOracle           = builtin.NameOracle
	AgentLibrarian        = builtin.NameLibrarian
	AgentExplore          = builtin.NameExplore
	AgentMultimodalLooker = builtin.NameMultimodalLooker
	AgentMetis            = builtin.NameMetis
	AgentMomus            = builtin.NameMomus
	AgentAtlas            = builtin.NameAtlas
	AgentSisyphusJunior   = builtin.NameSisyphusJunior
	AgentPrometheus       = builtin.NamePrometheus
)

// type alias for cleaner call sites.
type agentName = builtin.Name
