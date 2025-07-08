package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/decision"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// Store represents the memory storage system
type Store struct {
	db     *bolt.DB
	config *config.Config
}

// QueryRecord represents a stored query
type QueryRecord struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	UserIP    string    `json:"user_ip,omitempty"`
}

// ResponseRecord represents a stored response
type ResponseRecord struct {
	ID         string             `json:"id"`
	QueryID    string             `json:"query_id"`
	Response   string             `json:"response"`
	Analysis   *decision.Analysis `json:"analysis,omitempty"`
	Timestamp  time.Time          `json:"timestamp"`
}

// FeedbackRecord represents user feedback
type FeedbackRecord struct {
	ID        string    `json:"id"`
	QueryID   string    `json:"query_id"`
	Choice    string    `json:"choice"` // accept, decline, more_info
	Timestamp time.Time `json:"timestamp"`
}

// NewStore creates a new memory store
func NewStore(cfg *config.Config) (*Store, error) {
	// Ensure directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open BoltDB
	db, err := bolt.Open(cfg.DatabasePath, 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{
		db:     db,
		config: cfg,
	}

	// Initialize buckets
	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return store, nil
}

// initBuckets initializes database buckets
func (s *Store) initBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buckets := []string{"queries", "responses", "feedback", "metadata"}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
}

// StoreQuery stores a user query
func (s *Store) StoreQuery(query string) error {
	record := QueryRecord{
		ID:        generateID(),
		Query:     query,
		Timestamp: time.Now(),
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("queries"))
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(record.ID), data)
	})
}

// StoreResponse stores a response with analysis
func (s *Store) StoreResponse(query string, analysis *decision.Analysis) error {
	// Find the query ID (simplified - in production you'd want better query matching)
	queryID := generateID() // This should be the actual query ID

	record := ResponseRecord{
		ID:        generateID(),
		QueryID:   queryID,
		Response:  analysis.Response,
		Analysis:  analysis,
		Timestamp: time.Now(),
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("responses"))
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(record.ID), data)
	})
}

// StoreFeedback stores user feedback
func (s *Store) StoreFeedback(query, choice string) error {
	// Find the query ID (simplified)
	queryID := generateID() // This should be the actual query ID

	record := FeedbackRecord{
		ID:        generateID(),
		QueryID:   queryID,
		Choice:    choice,
		Timestamp: time.Now(),
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("feedback"))
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(record.ID), data)
	})
}

// GetRecentQueries retrieves recent queries
func (s *Store) GetRecentQueries(limit int) ([]QueryRecord, error) {
	var queries []QueryRecord

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("queries"))
		cursor := bucket.Cursor()

		count := 0
		for k, v := cursor.Last(); k != nil && count < limit; k, v = cursor.Prev() {
			var record QueryRecord
			if err := json.Unmarshal(v, &record); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal query record")
				continue
			}
			queries = append(queries, record)
			count++
		}

		return nil
	})

	return queries, err
}

// GetResponsesByQuery retrieves responses for a specific query
func (s *Store) GetResponsesByQuery(queryID string) ([]ResponseRecord, error) {
	var responses []ResponseRecord

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("responses"))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var record ResponseRecord
			if err := json.Unmarshal(v, &record); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal response record")
				continue
			}
			if record.QueryID == queryID {
				responses = append(responses, record)
			}
		}

		return nil
	})

	return responses, err
}

// GetFeedbackStats retrieves feedback statistics
func (s *Store) GetFeedbackStats() (map[string]int, error) {
	stats := make(map[string]int)

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("feedback"))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var record FeedbackRecord
			if err := json.Unmarshal(v, &record); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal feedback record")
				continue
			}
			stats[record.Choice]++
		}

		return nil
	})

	return stats, err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// generateID generates a simple ID based on timestamp
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
