package app

// Add additional model types for variable discovery, file analysis, etc.

// CandidateVariable represents a discovered variable in the source project
// for mapping to template variables.
type CandidateVariable struct {
	Name        string
	Occurrences int
	Example     string
}
