package model

// TagNode is one node of the five-root hierarchical tag tree.
//
// The tree is deliberately unbounded in depth: every node may have children,
// and children may have children, so that the Wails frontend can render the
// same recursive component at any drilling level.
type TagNode struct {
	// ID is the stable machine key, e.g. "mood.calm".
	ID string `json:"id"`
	// Root is the dimension this node belongs to.
	Root RootKey `json:"root"`
	// Parent is the id of the parent node ("" for roots).
	Parent string `json:"parent"`
	// Label is the Chinese display name.
	Label string `json:"label"`
	// Hint is a short clarifying phrase shown under the label.
	Hint string `json:"hint"`
	// Depth is 0 for the root nodes.
	Depth int `json:"depth"`
	// Children holds the ids of the child nodes.
	Children []string `json:"children,omitempty"`
	// Count is the number of verses matching this node alone.
	Count int `json:"count"`
	// Path is the full breadcrumb of labels from the root down to this node.
	Path []string `json:"path"`
}

// TreeNode is a tag node decorated with its live candidate count and the
// count of the verses that remain after the currently selected breadcrumb is
// applied.  It is what the frontend actually renders.
type TreeNode struct {
	TagNode
	// Remaining is the number of verses matching this node *within the
	// current selection*.
	Remaining int `json:"remaining"`
	// IsLeaf reports whether the node can be finalised.
	IsLeaf bool `json:"is_leaf"`
}

// Breadcrumb is one step of the drilling history.
type Breadcrumb struct {
	// NodeID is the selected tag node.
	NodeID string `json:"node_id"`
	// Label is the display label of the node.
	Label string `json:"label"`
	// Root is the dimension the node belongs to.
	Root RootKey `json:"root"`
	// Path is the full label path up to and including this node.
	Path []string `json:"path"`
	// Remaining is the candidate count after this step.
	Remaining int `json:"remaining"`
}

// MatchExplanation explains why a verse was chosen.
type MatchExplanation struct {
	// Overall is the composite score, 0-100.
	Overall int `json:"overall"`
	// Semantic is the artistic-conception resonance score.
	Semantic int `json:"semantic"`
	// Emotional is the emotional dimension fit score.
	Emotional int `json:"emotional"`
	// Scenario is the situational consistency score.
	Scenario int `json:"scenario"`
	// Temperament is the Dongpo temperament match score.
	Temperament int `json:"temperament"`
	// Reasons lists the human-readable justifications.
	Reasons []string `json:"reasons"`
	// MatchedTags lists the tags that contributed to the score.
	MatchedTags []string `json:"matched_tags"`
}

// MatchResult is the output of the final convergence matching rule.
type MatchResult struct {
	// Primary is the single most resonant verse or couplet.
	Primary Verse `json:"primary"`
	// Alternatives are 2-3 secondary suggestions.
	Alternatives []Verse `json:"alternatives"`
	// Explanation explains the primary match.
	Explanation MatchExplanation `json:"explanation"`
	// Candidates is the number of verses that were considered.
	Candidates int `json:"candidates"`
}

// Stats summarises the loaded corpus.
type Stats struct {
	Poems    int `json:"poems"`
	Verses   int `json:"verses"`
	Tunes    int `json:"tunes"`
	Lines    int `json:"lines"`
	Couplets int `json:"couplets"`
	Concepts int `json:"concepts"`
}
