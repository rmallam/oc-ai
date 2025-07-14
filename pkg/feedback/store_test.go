package feedback

import (
	"os"
	"testing"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/models"
	bolt "go.etcd.io/bbolt"
)

func TestStore_SaveAndGetFeedback(t *testing.T) {
	// Create a temporary database for testing
	db, err := bolt.Open("test.db", 0600, nil)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer os.Remove("test.db")
	defer db.Close()

	// Create a new feedback store
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("failed to create feedback store: %v", err)
	}

	// Create a sample analysis
	analysis := &models.Analysis{
		Query:    "test query",
		Response: "test response",
	}

	// Save the feedback
	if err := store.SaveFeedback("test query", analysis); err != nil {
		t.Fatalf("failed to save feedback: %v", err)
	}

	// Get the feedback
	retrievedAnalysis, err := store.GetFeedback("test query")
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	// Check if the retrieved analysis is correct
	if retrievedAnalysis.Query != "test query" {
		t.Errorf("expected query to be 'test query', got '%s'", retrievedAnalysis.Query)
	}
	if retrievedAnalysis.Response != "test response" {
		t.Errorf("expected response to be 'test response', got '%s'", retrievedAnalysis.Response)
	}
}
