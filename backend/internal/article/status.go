package article

// Status is the article lifecycle state.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

// allowedTransitions defines the article state machine: which target states are
// reachable from each source state. It is the single source of truth for what
// lifecycle moves are legal; operations enforce the richer business rules on top
// of it (see operation/publish).
var allowedTransitions = map[Status][]Status{
	StatusDraft:     {StatusPublished, StatusArchived},
	StatusPublished: {StatusArchived},
	StatusArchived:  {},
}

// CanTransition reports whether moving from -> to is a legal state change.
func CanTransition(from, to Status) bool {
	for _, candidate := range allowedTransitions[from] {
		if candidate == to {
			return true
		}
	}
	return false
}
