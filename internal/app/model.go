package app

import (
	"time"
)

// Add additional model types for variable discovery, file analysis, etc.

// CandidateVariable represents a discovered variable in the source project
// for mapping to template variables.
type CandidateVariable struct {
	Name        string
	Occurrences int
	Example     string
}

type templateMetadata struct {
	Token       map[string]string   `json:"token"`
	Variables   map[string]Variable `json:"variables"`
	StaticFiles []string            `json:"staticFiles"`
	GeneratedAt time.Time           `json:"generatedAt"`
}
