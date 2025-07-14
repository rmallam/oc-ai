package decision

import (
	"regexp"
	"strings"
)

// isOperatorQuery checks if the prompt is asking about an operator's presence
func isOperatorQuery(prompt string) (bool, string) {
	// Example: "is gatekeeper operator installed", "is prometheus operator present"
	// Returns (true, "gatekeeper")
	pattern := regexp.MustCompile(`is ([a-z0-9-]+) operator (installed|present|running|enabled|deployed)`)
	matches := pattern.FindStringSubmatch(strings.ToLower(prompt))
	if len(matches) > 1 {
		return true, matches[1]
	}
	return false, ""
}
