package aiadvisor

import (
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Conversation represents a conversation session
type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QueryRequest represents an incoming query request
type QueryRequest struct {
	Query          string `json:"query" binding:"required"`
	ConversationID string `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id" binding:"required"`
}

// QueryResponse represents the response to a query
type QueryResponse struct {
	Answer         string   `json:"answer"`
	ConversationID string   `json:"conversation_id"`
	Sources        []string `json:"sources,omitempty"`
	Suggestions    []string `json:"suggestions,omitempty"`
}

// RecommendationRequest represents a request for recommendations
type RecommendationRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Limit  int    `json:"limit,omitempty"`
}

// Recommendation represents a recommended feature or action
type Recommendation struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Relevance   float64 `json:"relevance"`
	Action      string  `json:"action,omitempty"`
}

// DiagnosticRequest represents a diagnostic request
type DiagnosticRequest struct {
	UserID      string   `json:"user_id" binding:"required"`
	Symptoms    []string `json:"symptoms" binding:"required"`
	SystemInfo  string   `json:"system_info,omitempty"`
}

// DiagnosticResult represents a diagnostic result
type DiagnosticResult struct {
	Issue       string   `json:"issue"`
	Severity    string   `json:"severity"` // "low", "medium", "high", "critical"
	Causes      []string `json:"causes"`
	Solutions   []string `json:"solutions"`
	References  []string `json:"references,omitempty"`
}

// KnowledgeArticle represents a knowledge base article
type KnowledgeArticle struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// SystemStatus represents the current system status
type SystemStatus struct {
	Status       string `json:"status"`
	KnowledgeOK  bool   `json:"knowledge_ok"`
	LLMOK        bool   `json:"llm_ok"`
	ActiveConvos int    `json:"active_conversations"`
	Uptime       string `json:"uptime"`
}

// HealthCheck represents a health check response
type HealthCheck struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
