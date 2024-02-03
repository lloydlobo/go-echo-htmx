package services

type (
	Filter struct {
		Url      string
		Name     string
		Selected bool
	}

	// Action implements enumeration of actions
	Action int
)

// Enumerate Action related constants in one type
const (
	ActionCreate Action = iota
	ActionToggle
	ActionEdit
	ActionUpdate
	ActionDelete
)
