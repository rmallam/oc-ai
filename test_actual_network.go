package main

import (
	"fmt"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/network"
)

func main() {
	engine := network.NewTroubleshootingEngine()
	query := "create a test namespace and a service account called test-sa, this service account should have admin access to test namespace"

	fmt.Printf("Testing query: %s\n", query)
	isNetworkQuery := engine.IsNetworkQuery(query)
	fmt.Printf("IsNetworkQuery result: %t\n", isNetworkQuery)

	if isNetworkQuery {
		fmt.Println("WARNING: This query is incorrectly being classified as a network troubleshooting query!")
	} else {
		fmt.Println("Good: This query is NOT being classified as a network query.")
	}
}
