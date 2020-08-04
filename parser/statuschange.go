package parser

// StatusChange is a change request against the current processing stats.
type StatusChange struct {
	LayerName string
	Status    string
}
