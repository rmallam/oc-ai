package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/types"
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
	ID        string          `json:"id"`
	QueryID   string          `json:"query_id"`
	Response  string          `json:"response"`
	Analysis  *types.Analysis `json:"analysis,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// FeedbackRecord represents user feedback
type FeedbackRecord struct {
	ID        string    `json:"id"`
	QueryID   string    `json:"query_id"`
	Choice    string    `json:"choice"` // accept, decline, more_info
	Timestamp time.Time `json:"timestamp"`
}

// PromptCategory represents a categorized prompt
type PromptCategory struct {
	ID          string    `json:"id"`
	Prompt      string    `json:"prompt"`
	Category    string    `json:"category"`
	Subcategory string    `json:"subcategory"`
	Frequency   int       `json:"frequency"`
	LastUsed    time.Time `json:"last_used"`
	Success     bool      `json:"success"`
	Confidence  float64   `json:"confidence"`
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
		buckets := []string{"queries", "responses", "feedback", "metadata", "prompt_categories"}
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
func (s *Store) StoreResponse(query string, analysis *types.Analysis) error {
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

// StorePromptCategory stores or updates a categorized prompt
func (s *Store) StorePromptCategory(prompt string, category string, subcategory string, success bool, confidence float64) error {
	// Check if prompt already exists
	existingRecord, err := s.getPromptCategory(prompt)
	if err != nil && err.Error() != "prompt not found" {
		return fmt.Errorf("failed to check existing prompt: %w", err)
	}

	var record PromptCategory
	if existingRecord != nil {
		// Update existing record
		record = *existingRecord
		record.Frequency++
		record.LastUsed = time.Now()
		record.Success = success
		record.Confidence = confidence
	} else {
		// Create new record
		record = PromptCategory{
			ID:          generateID(),
			Prompt:      prompt,
			Category:    category,
			Subcategory: subcategory,
			Frequency:   1,
			LastUsed:    time.Now(),
			Success:     success,
			Confidence:  confidence,
		}
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("prompt_categories"))
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal prompt category: %w", err)
		}
		return bucket.Put([]byte(record.ID), data)
	})
}

// getPromptCategory retrieves a prompt category by prompt text
func (s *Store) getPromptCategory(prompt string) (*PromptCategory, error) {
	var record *PromptCategory

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("prompt_categories"))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var pc PromptCategory
			if err := json.Unmarshal(v, &pc); err != nil {
				continue
			}
			if pc.Prompt == prompt {
				record = &pc
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("prompt not found")
	}
	return record, nil
}

// GetPromptCategories retrieves all prompt categories
func (s *Store) GetPromptCategories() ([]PromptCategory, error) {
	var categories []PromptCategory

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("prompt_categories"))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var record PromptCategory
			if err := json.Unmarshal(v, &record); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal prompt category")
				continue
			}
			categories = append(categories, record)
		}
		return nil
	})

	return categories, err
}

// UpdatePromptsFile updates the prompts.md file with new categorized prompts
func (s *Store) UpdatePromptsFile() error {
	categories, err := s.GetPromptCategories()
	if err != nil {
		return fmt.Errorf("failed to get prompt categories: %w", err)
	}

	// Group by category and subcategory
	grouped := make(map[string]map[string][]PromptCategory)
	for _, pc := range categories {
		if grouped[pc.Category] == nil {
			grouped[pc.Category] = make(map[string][]PromptCategory)
		}
		grouped[pc.Category][pc.Subcategory] = append(grouped[pc.Category][pc.Subcategory], pc)
	}

	// Build the new prompts section
	var newPrompts strings.Builder
	newPrompts.WriteString("\n---\n\n## User-Generated Prompts\n\n")
	newPrompts.WriteString("*These prompts were automatically collected from user interactions and usage patterns.*\n\n")

	// Sort categories for consistent output
	var sortedCategories []string
	for cat := range grouped {
		sortedCategories = append(sortedCategories, cat)
	}
	sort.Strings(sortedCategories)

	for _, category := range sortedCategories {
		newPrompts.WriteString(fmt.Sprintf("### %s\n\n", category))

		// Sort subcategories
		var sortedSubcategories []string
		for subcat := range grouped[category] {
			sortedSubcategories = append(sortedSubcategories, subcat)
		}
		sort.Strings(sortedSubcategories)

		for _, subcategory := range sortedSubcategories {
			if subcategory != "" {
				newPrompts.WriteString(fmt.Sprintf("#### %s\n", subcategory))
			}

			// Sort prompts by frequency (descending)
			prompts := grouped[category][subcategory]
			sort.Slice(prompts, func(i, j int) bool {
				return prompts[i].Frequency > prompts[j].Frequency
			})

			for _, prompt := range prompts {
				// Show frequency and confidence for popular prompts
				if prompt.Frequency > 1 {
					newPrompts.WriteString(fmt.Sprintf("- \"%s\" *(used %d times, %.0f%% confidence)*\n",
						prompt.Prompt, prompt.Frequency, prompt.Confidence*100))
				} else {
					newPrompts.WriteString(fmt.Sprintf("- \"%s\"\n", prompt.Prompt))
				}
			}
			newPrompts.WriteString("\n")
		}
	}

	// Read the current prompts.md file
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	promptsFile := filepath.Join(currentDir, "prompts.md")

	content, err := os.ReadFile(promptsFile)
	if err != nil {
		return fmt.Errorf("failed to read prompts.md: %w", err)
	}

	// Remove existing user-generated section if it exists
	contentStr := string(content)
	startMarker := "\n---\n\n## User-Generated Prompts\n"
	if idx := strings.Index(contentStr, startMarker); idx != -1 {
		contentStr = contentStr[:idx]
	}

	// Append new user-generated prompts
	updatedContent := contentStr + newPrompts.String()

	// Write back to file
	if err := os.WriteFile(promptsFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write prompts.md: %w", err)
	}

	logrus.Infof("Updated prompts.md with %d user-generated prompts", len(categories))
	return nil
}

// CategorizePrompt automatically categorizes a prompt based on keywords
func (s *Store) CategorizePrompt(prompt string) (category string, subcategory string) {
	lowerPrompt := strings.ToLower(prompt)

	// Define category mapping
	categoryMap := map[string]map[string][]string{
		"Cluster Administration": {
			"Basic Information":    {"cluster", "version", "info", "capacity", "operators"},
			"Namespace Management": {"namespace", "create namespace", "delete namespace"},
			"Resource Quotas":      {"quota", "limits", "resource usage"},
			"Maintenance":          {"upgrade", "update", "maintenance", "backup"},
		},
		"SRE Tasks": {
			"Health Checks":     {"health", "unhealthy", "failing", "alerts"},
			"Performance":       {"cpu", "memory", "disk", "performance", "top", "usage"},
			"Incident Response": {"crash", "error", "failed", "pending", "evicted", "oom"},
			"Capacity Planning": {"utilization", "capacity", "resources", "growth"},
		},
		"Application Deployment": {
			"Deployment Management": {"deploy", "scale", "rollback", "rollout"},
			"Helm Charts":           {"helm", "install", "upgrade", "chart", "release"},
			"Pod Management":        {"pod", "logs", "exec", "describe", "port-forward"},
			"Service Management":    {"service", "expose", "endpoints", "connectivity"},
		},
		"Networking": {
			"Service Discovery": {"service", "endpoints", "dns", "mesh"},
			"Ingress & Routes":  {"ingress", "route", "ssl", "certificate"},
			"Network Policies":  {"network policy", "traffic", "flows"},
			"Load Balancing":    {"load balancer", "lb", "loadbalancer"},
		},
		"Storage": {
			"Persistent Volumes": {"pv", "pvc", "persistent", "volume", "storage"},
			"Storage Classes":    {"storage class", "provisioner"},
			"Backup & Recovery":  {"backup", "snapshot", "restore"},
		},
		"Security": {
			"RBAC":             {"rbac", "role", "permission", "access", "service account"},
			"Security Context": {"security", "privileged", "policy"},
			"Secrets":          {"secret", "certificate", "key", "password"},
		},
		"Monitoring": {
			"Metrics":  {"metrics", "prometheus", "grafana"},
			"Logging":  {"logs", "logging", "audit"},
			"Alerting": {"alert", "notification", "alarm"},
		},
		"Troubleshooting": {
			"Pod Issues":         {"why", "troubleshoot", "diagnose", "fix", "check"},
			"Network Issues":     {"connectivity", "network", "dns resolution"},
			"Storage Issues":     {"mount", "volume", "disk"},
			"Performance Issues": {"slow", "latency", "leak"},
		},
		"Node Management": {
			"Node Operations":      {"node", "drain", "cordon", "uncordon"},
			"Node Troubleshooting": {"node not ready", "kubelet", "resource pressure"},
			"Scheduling":           {"schedule", "affinity", "taint", "toleration"},
		},
		"CI/CD": {
			"Pipeline":         {"pipeline", "build", "deployment"},
			"Image Management": {"image", "registry", "scan"},
			"GitOps":           {"git", "argocd", "sync"},
		},
	}

	// Find the best category match
	bestCategory := "Troubleshooting"
	bestSubcategory := "General"
	maxMatches := 0

	for cat, subcats := range categoryMap {
		for subcat, keywords := range subcats {
			matches := 0
			for _, keyword := range keywords {
				if strings.Contains(lowerPrompt, keyword) {
					matches++
				}
			}
			if matches > maxMatches {
				maxMatches = matches
				bestCategory = cat
				bestSubcategory = subcat
			}
		}
	}

	return bestCategory, bestSubcategory
}
