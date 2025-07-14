package main

import (
	"fmt"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/executor"
)

func main() {
	executor := executor.NewCommandExecutor()

	// Test simple commands first
	commands := []string{
		"echo hello",
		"echo hello && echo world",
		"kubectl version --client",
		"kubectl create namespace testing && kubectl create serviceaccount test-sa -n testing",
	}

	for i, command := range commands {
		fmt.Printf("\n=== Test %d ===\n", i+1)
		fmt.Printf("Command: %s\n", command)
		fmt.Printf("Is command safe: %t\n", executor.IsCommandSafe(command))

		result := executor.Execute(command)

		fmt.Printf("Exit code: %d\n", result.ExitCode)
		fmt.Printf("Output: %s\n", result.Output)
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}
}
