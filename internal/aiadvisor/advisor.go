package aiadvisor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// AdvisorConfig holds configuration for the AI advisor
type AdvisorConfig struct {
	LLM              LLMConfig `json:"llm"`
	KnowledgePath    string    `json:"knowledge_path"`
	MaxMessages      int       `json:"max_messages"`
	CleanupInterval  int       `json:"cleanup_interval"`
	MaxConversationAge int     `json:"max_conversation_age"`
}

// Advisor represents the AI advisor service
type Advisor struct {
	config    AdvisorConfig
	llm       *LLMClient
	knowledge *KnowledgeBase
	conversations *ConversationManager
	startTime time.Time
}

// NewAdvisor creates a new AI advisor
func NewAdvisor(config AdvisorConfig) *Advisor {
	llm := NewLLMClient(config.LLM)
	knowledge := NewKnowledgeBase(config.KnowledgePath)
	conversations := NewConversationManager(config.MaxMessages)

	return &Advisor{
		config:        config,
		llm:           llm,
		knowledge:     knowledge,
		conversations: conversations,
		startTime:     time.Now(),
	}
}

// Initialize initializes the advisor
func (a *Advisor) Initialize() error {
	// Load knowledge base
	if err := a.knowledge.Load(); err != nil {
		return fmt.Errorf("failed to load knowledge base: %w", err)
	}

	log.Printf("Knowledge base loaded with %d articles", a.knowledge.Count())

	// Start cleanup goroutine
	go a.cleanupLoop()

	return nil
}

// cleanupLoop periodically cleans up old conversations
func (a *Advisor) cleanupLoop() {
	ticker := time.NewTicker(time.Duration(a.config.CleanupInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		maxAge := time.Duration(a.config.MaxConversationAge) * time.Hour
		count := a.conversations.Cleanup(maxAge)
		if count > 0 {
			log.Printf("Cleaned up %d old conversations", count)
		}
	}
}

// Query handles a user query
func (a *Advisor) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// Get or create conversation
	conversation := a.conversations.GetOrCreate(req.ConversationID, req.UserID)

	// Add user message to conversation
	if _, err := a.conversations.AddMessage(conversation.ID, "user", req.Query); err != nil {
		return nil, fmt.Errorf("failed to add user message: %w", err)
	}

	// Search knowledge base for relevant articles
	articles := a.knowledge.Search(req.Query)
	knowledgeContext := a.buildKnowledgeContext(articles)

	// Build system prompt
	systemPrompt := a.buildSystemPrompt(knowledgeContext)

	// Get conversation messages for LLM
	messages, err := a.conversations.ToLLMMessages(conversation.ID, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Generate response from LLM
	answer, err := a.llm.GenerateResponse(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Add assistant message to conversation
	if _, err := a.conversations.AddMessage(conversation.ID, "assistant", answer); err != nil {
		return nil, fmt.Errorf("failed to add assistant message: %w", err)
	}

	// Extract sources from articles
	sources := make([]string, 0, len(articles))
	for _, article := range articles {
		sources = append(sources, article.Title)
	}

	// Generate suggestions
	suggestions := a.generateSuggestions(req.Query, articles)

	return &QueryResponse{
		Answer:         answer,
		ConversationID: conversation.ID,
		Sources:        sources,
		Suggestions:    suggestions,
	}, nil
}

// buildKnowledgeContext builds context from knowledge base articles
func (a *Advisor) buildKnowledgeContext(articles []*KnowledgeArticle) string {
	if len(articles) == 0 {
		return ""
	}

	var context strings.Builder
	context.WriteString("Relevant knowledge base articles:\n\n")

	for i, article := range articles {
		if i >= 5 { // Limit to top 5 articles
			break
		}
		context.WriteString(fmt.Sprintf("## %s\n%s\n\n", article.Title, article.Content))
	}

	return context.String()
}

// buildSystemPrompt builds the system prompt for the LLM
func (a *Advisor) buildSystemPrompt(knowledgeContext string) string {
	prompt := `You are an AI advisor for a NAS operating system. Your role is to help users with:

1. Answering questions about system features and configuration
2. Providing guidance on common tasks
3. Helping diagnose and solve problems
4. Recommending features based on user needs

Guidelines:
- Be helpful, clear, and concise
- Use the knowledge base information when available
- If you don't know the answer, say so honestly
- Provide step-by-step instructions when appropriate
- Use markdown formatting for better readability

`

	if knowledgeContext != "" {
		prompt += knowledgeContext
	}

	return prompt
}

// generateSuggestions generates follow-up suggestions
func (a *Advisor) generateSuggestions(query string, articles []*KnowledgeArticle) []string {
	suggestions := make([]string, 0)

	// Add suggestions based on article categories
	categories := make(map[string]bool)
	for _, article := range articles {
		categories[article.Category] = true
	}

	for category := range categories {
		switch category {
		case "storage":
			suggestions = append(suggestions, "How to manage storage pools?")
			suggestions = append(suggestions, "What RAID configuration should I use?")
		case "network":
			suggestions = append(suggestions, "How to set up remote access?")
			suggestions = append(suggestions, "How to configure firewall rules?")
		case "backup":
			suggestions = append(suggestions, "How to set up automatic backups?")
			suggestions = append(suggestions, "What backup strategy is recommended?")
		case "docker":
			suggestions = append(suggestions, "How to install Docker containers?")
			suggestions = append(suggestions, "How to manage Docker volumes?")
		}
	}

	// Limit suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}

// GetRecommendations gets recommendations for a user
func (a *Advisor) GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]Recommendation, error) {
	// Get user's conversation history
	conversations := a.conversations.GetUserConversations(req.UserID)

	// Analyze user's interests from conversation history
	interests := a.analyzeInterests(conversations)

	// Get recommendations based on interests
	recommendations := a.generateRecommendations(interests)

	// Limit recommendations
	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

// analyzeInterests analyzes user interests from conversation history
func (a *Advisor) analyzeInterests(conversations []*Conversation) map[string]int {
	interests := make(map[string]int)

	for _, conversation := range conversations {
		for _, message := range conversation.Messages {
			if message.Role == "user" {
				// Simple keyword analysis
				content := strings.ToLower(message.Content)
				if strings.Contains(content, "storage") || strings.Contains(content, "disk") || strings.Contains(content, "raid") {
					interests["storage"]++
				}
				if strings.Contains(content, "network") || strings.Contains(content, "remote") || strings.Contains(content, "vpn") {
					interests["network"]++
				}
				if strings.Contains(content, "backup") || strings.Contains(content, "snapshot") || strings.Contains(content, "restore") {
					interests["backup"]++
				}
				if strings.Contains(content, "docker") || strings.Contains(content, "container") || strings.Contains(content, "app") {
					interests["docker"]++
				}
				if strings.Contains(content, "security") || strings.Contains(content, "firewall") || strings.Contains(content, "permission") {
					interests["security"]++
				}
			}
		}
	}

	return interests
}

// generateRecommendations generates recommendations based on interests
func (a *Advisor) generateRecommendations(interests map[string]int) []Recommendation {
	recommendations := make([]Recommendation, 0)

	// Storage recommendations
	if count, ok := interests["storage"]; ok && count > 0 {
		recommendations = append(recommendations, Recommendation{
			ID:          "rec-storage-pool",
			Title:       "Optimize Storage Pool",
			Description: "Based on your interest in storage, consider optimizing your storage pool configuration for better performance.",
			Category:    "storage",
			Relevance:   float64(count) / 10.0,
			Action:      "Navigate to Storage Manager",
		})
	}

	// Network recommendations
	if count, ok := interests["network"]; ok && count > 0 {
		recommendations = append(recommendations, Recommendation{
			ID:          "rec-remote-access",
			Title:       "Set Up Remote Access",
			Description: "Enable secure remote access to your NAS from anywhere.",
			Category:    "network",
			Relevance:   float64(count) / 10.0,
			Action:      "Configure Remote Access",
		})
	}

	// Backup recommendations
	if count, ok := interests["backup"]; ok && count > 0 {
		recommendations = append(recommendations, Recommendation{
			ID:          "rec-auto-backup",
			Title:       "Configure Automatic Backups",
			Description: "Set up automatic backups to protect your important data.",
			Category:    "backup",
			Relevance:   float64(count) / 10.0,
			Action:      "Set Up Backup Schedule",
		})
	}

	// Docker recommendations
	if count, ok := interests["docker"]; ok && count > 0 {
		recommendations = append(recommendations, Recommendation{
			ID:          "rec-docker-apps",
			Title:       "Popular Docker Applications",
			Description: "Discover and install popular Docker applications for your NAS.",
			Category:    "docker",
			Relevance:   float64(count) / 10.0,
			Action:      "Browse Docker Apps",
		})
	}

	// Security recommendations
	if count, ok := interests["security"]; ok && count > 0 {
		recommendations = append(recommendations, Recommendation{
			ID:          "rec-security-audit",
			Title:       "Security Audit",
			Description: "Run a security audit to identify and fix potential vulnerabilities.",
			Category:    "security",
			Relevance:   float64(count) / 10.0,
			Action:      "Run Security Audit",
		})
	}

	// Add general recommendations if no specific interests
	if len(recommendations) == 0 {
		recommendations = append(recommendations, Recommendation{
			ID:          "rec-getting-started",
			Title:       "Getting Started Guide",
			Description: "New to NAS? Start with our comprehensive getting started guide.",
			Category:    "general",
			Relevance:   1.0,
			Action:      "View Getting Started Guide",
		})
	}

	return recommendations
}

// Diagnose performs system diagnostics
func (a *Advisor) Diagnose(ctx context.Context, req *DiagnosticRequest) (*DiagnosticResult, error) {
	// Build diagnostic prompt
	prompt := a.buildDiagnosticPrompt(req)

	// Create a temporary conversation for diagnostic
	conversation := a.conversations.Create(req.UserID)
	defer a.conversations.Delete(conversation.ID)

	// Add diagnostic prompt
	if _, err := a.conversations.AddMessage(conversation.ID, "user", prompt); err != nil {
		return nil, fmt.Errorf("failed to add diagnostic prompt: %w", err)
	}

	// Get conversation messages
	messages, err := a.conversations.ToLLMMessages(conversation.ID, a.getDiagnosticSystemPrompt())
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Generate diagnostic result
	response, err := a.llm.GenerateResponse(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diagnostic: %w", err)
	}

	// Parse diagnostic result
	result := a.parseDiagnosticResponse(response, req)

	return result, nil
}

// buildDiagnosticPrompt builds a diagnostic prompt
func (a *Advisor) buildDiagnosticPrompt(req *DiagnosticRequest) string {
	prompt := fmt.Sprintf("Please diagnose the following issue:\n\nSymptoms: %s\n", strings.Join(req.Symptoms, ", "))

	if req.SystemInfo != "" {
		prompt += fmt.Sprintf("\nSystem Information:\n%s", req.SystemInfo)
	}

	prompt += "\n\nPlease provide:\n1. The likely issue\n2. Possible causes\n3. Recommended solutions"

	return prompt
}

// getDiagnosticSystemPrompt returns the system prompt for diagnostics
func (a *Advisor) getDiagnosticSystemPrompt() string {
	return `You are a NAS system diagnostic expert. Analyze the provided symptoms and system information to:

1. Identify the most likely issue
2. List possible causes
3. Provide step-by-step solutions

Format your response as a structured diagnostic report. Be specific and actionable.

Consider common NAS issues:
- Storage failures (disk errors, RAID degradation)
- Network connectivity problems
- Performance issues
- Permission and access problems
- Backup failures
- Docker container issues`
}

// parseDiagnosticResponse parses the LLM response into a diagnostic result
func (a *Advisor) parseDiagnosticResponse(response string, req *DiagnosticRequest) *DiagnosticResult {
	// Simple parsing - in production, use more sophisticated parsing
	result := &DiagnosticResult{
		Issue:     "System diagnostic analysis",
		Severity:  "medium",
		Causes:    []string{"See detailed analysis below"},
		Solutions: []string{response},
	}

	// Try to determine severity based on symptoms
	for _, symptom := range req.Symptoms {
		symptom = strings.ToLower(symptom)
		if strings.Contains(symptom, "critical") || strings.Contains(symptom, "failure") || strings.Contains(symptom, "error") {
			result.Severity = "high"
			break
		}
		if strings.Contains(symptom, "slow") || strings.Contains(symptom, "performance") {
			result.Severity = "medium"
			break
		}
	}

	return result
}

// GetStatus returns the current status of the advisor
func (a *Advisor) GetStatus() *SystemStatus {
	llmOK := true
	if err := a.llm.HealthCheck(context.Background()); err != nil {
		llmOK = false
	}

	return &SystemStatus{
		Status:       "running",
		KnowledgeOK:  a.knowledge.Count() > 0,
		LLMOK:        llmOK,
		ActiveConvos: a.conversations.Count(),
		Uptime:       time.Since(a.startTime).String(),
	}
}

// HealthCheck performs a health check
func (a *Advisor) HealthCheck() *HealthCheck {
	return &HealthCheck{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
}
