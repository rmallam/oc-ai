package models

import "time"

// Analysis represents the result of analysis
type Analysis struct {
	Query       string                 `json:"query"`
	Response    string                 `json:"response"`
	Confidence  float64                `json:"confidence"`
	Severity    string                 `json:"severity"`
	RootCauses  []RootCause            `json:"root_causes"`
	Actions     []RecommendedAction    `json:"recommended_actions"`
	Evidence    []Evidence             `json:"evidence"`
	Timestamp   time.Time              `json:"timestamp"`
	AnalysisID  string                 `json:"analysis_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RootCause represents an identified root cause
type RootCause struct {
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Evidence    string  `json:"evidence"`
}

// RecommendedAction represents a recommended action
type RecommendedAction struct {
	Description string `json:"description"`
	Priority    string `json:"priority"` // High, Medium, Low
	Command     string `json:"command,omitempty"`
	Risk        string `json:"risk,omitempty"`
}

// Evidence represents evidence collected during analysis
type Evidence struct {
	Type        string    `json:"type"`        // logs, events, status, etc.
	Source      string    `json:"source"`      // pod name, node name, etc.
	Content     string    `json:"content"`     // actual evidence content
	Timestamp   time.Time `json:"timestamp"`
}
