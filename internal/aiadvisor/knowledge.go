package aiadvisor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// KnowledgeBase manages the knowledge base articles
type KnowledgeBase struct {
	mu       sync.RWMutex
	articles map[string]*KnowledgeArticle
	basePath string
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase(basePath string) *KnowledgeBase {
	return &KnowledgeBase{
		articles: make(map[string]*KnowledgeArticle),
		basePath: basePath,
	}
}

// Load loads articles from the knowledge base directory
func (kb *KnowledgeBase) Load() error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	return filepath.Walk(kb.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		var article KnowledgeArticle
		if err := json.Unmarshal(data, &article); err != nil {
			return fmt.Errorf("failed to unmarshal article %s: %w", path, err)
		}

		kb.articles[article.ID] = &article
		return nil
	})
}

// Get retrieves an article by ID
func (kb *KnowledgeBase) Get(id string) (*KnowledgeArticle, bool) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	article, ok := kb.articles[id]
	return article, ok
}

// Search searches for articles matching the query
func (kb *KnowledgeBase) Search(query string) []*KnowledgeArticle {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*KnowledgeArticle

	for _, article := range kb.articles {
		if matchesQuery(article, query) {
			results = append(results, article)
		}
	}

	return results
}

// matchesQuery checks if an article matches the search query
func matchesQuery(article *KnowledgeArticle, query string) bool {
	// Check title
	if strings.Contains(strings.ToLower(article.Title), query) {
		return true
	}

	// Check content
	if strings.Contains(strings.ToLower(article.Content), query) {
		return true
	}

	// Check tags
	for _, tag := range article.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Check category
	if strings.Contains(strings.ToLower(article.Category), query) {
		return true
	}

	return false
}

// GetByCategory retrieves articles by category
func (kb *KnowledgeBase) GetByCategory(category string) []*KnowledgeArticle {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	category = strings.ToLower(category)
	var results []*KnowledgeArticle

	for _, article := range kb.articles {
		if strings.ToLower(article.Category) == category {
			results = append(results, article)
		}
	}

	return results
}

// GetAll retrieves all articles
func (kb *KnowledgeBase) GetAll() []*KnowledgeArticle {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	articles := make([]*KnowledgeArticle, 0, len(kb.articles))
	for _, article := range kb.articles {
		articles = append(articles, article)
	}

	return articles
}

// Add adds a new article to the knowledge base
func (kb *KnowledgeBase) Add(article *KnowledgeArticle) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	if article.ID == "" {
		return fmt.Errorf("article ID cannot be empty")
	}

	// Save to file
	filePath := filepath.Join(kb.basePath, article.ID+".json")
	data, err := json.MarshalIndent(article, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal article: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write article file: %w", err)
	}

	kb.articles[article.ID] = article
	return nil
}

// Delete deletes an article from the knowledge base
func (kb *KnowledgeBase) Delete(id string) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	if _, ok := kb.articles[id]; !ok {
		return fmt.Errorf("article %s not found", id)
	}

	// Delete file
	filePath := filepath.Join(kb.basePath, id+".json")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete article file: %w", err)
	}

	delete(kb.articles, id)
	return nil
}

// Count returns the number of articles in the knowledge base
func (kb *KnowledgeBase) Count() int {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	return len(kb.articles)
}

// GetCategories returns all unique categories
func (kb *KnowledgeBase) GetCategories() []string {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	categories := make(map[string]bool)
	for _, article := range kb.articles {
		categories[article.Category] = true
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}

	return result
}
