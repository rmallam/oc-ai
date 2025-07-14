package feedback

import (
	"encoding/json"
	"fmt"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/models"
	bolt "go.etcd.io/bbolt"
)

// Store handles the storage and retrieval of user feedback
type Store struct {
	db *bolt.DB
}

// NewStore creates a new feedback store
func NewStore(db *bolt.DB) (*Store, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("feedback"))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback bucket: %w", err)
	}
	return &Store{db: db}, nil
}

// SaveFeedback saves a given analysis as positive feedback
func (s *Store) SaveFeedback(query string, analysis *models.Analysis) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("feedback"))
		key := []byte(query)
		val, err := json.Marshal(analysis)
		if err != nil {
			return fmt.Errorf("failed to marshal analysis: %w", err)
		}
		return b.Put(key, val)
	})
}

// GetFeedback retrieves feedback for a given query
func (s *Store) GetFeedback(query string) (*models.Analysis, error) {
	var analysis models.Analysis
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("feedback"))
		val := b.Get([]byte(query))
		if val == nil {
			return nil // No feedback found
		}
		return json.Unmarshal(val, &analysis)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}
	if analysis.Query == "" {
		return nil, nil // No feedback found
	}
	return &analysis, nil
}
