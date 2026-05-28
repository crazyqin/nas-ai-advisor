package aiadvisor

import (
	"os"
	"testing"
	"time"
)

// Helper function to create temp directory for tests
func createTempKnowledgeDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "knowledge-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func TestNewAdvisor(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)
	if advisor == nil {
		t.Fatal("Expected advisor to be created")
	}

	if advisor.config.LLM.Model != "llama2" {
		t.Errorf("Expected model to be llama2, got %s", advisor.config.LLM.Model)
	}
}

func TestConversationManager(t *testing.T) {
	cm := NewConversationManager(10)

	// Test Create
	conversation := cm.Create("user1")
	if conversation == nil {
		t.Fatal("Expected conversation to be created")
	}

	if conversation.UserID != "user1" {
		t.Errorf("Expected user ID to be user1, got %s", conversation.UserID)
	}

	// Test Get
	retrieved, ok := cm.Get(conversation.ID)
	if !ok {
		t.Fatal("Expected conversation to be found")
	}

	if retrieved.ID != conversation.ID {
		t.Errorf("Expected conversation ID to be %s, got %s", conversation.ID, retrieved.ID)
	}

	// Test AddMessage
	msg, err := cm.AddMessage(conversation.ID, "user", "Hello")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	if msg.Role != "user" {
		t.Errorf("Expected role to be user, got %s", msg.Role)
	}

	// Test GetMessages
	messages, err := cm.GetMessages(conversation.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Test GetUserConversations
	conversations := cm.GetUserConversations("user1")
	if len(conversations) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(conversations))
	}

	// Test Delete
	err = cm.Delete(conversation.ID)
	if err != nil {
		t.Fatalf("Failed to delete conversation: %v", err)
	}

	_, ok = cm.Get(conversation.ID)
	if ok {
		t.Error("Expected conversation to be deleted")
	}
}

func TestConversationManager_MaxMessages(t *testing.T) {
	cm := NewConversationManager(2)

	conversation := cm.Create("user1")

	// Add 3 messages
	cm.AddMessage(conversation.ID, "user", "Message 1")
	cm.AddMessage(conversation.ID, "assistant", "Response 1")
	cm.AddMessage(conversation.ID, "user", "Message 2")

	messages, _ := cm.GetMessages(conversation.ID)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages (max), got %d", len(messages))
	}

	// Verify the last 2 messages are kept
	if messages[0].Content != "Response 1" {
		t.Errorf("Expected first message to be 'Response 1', got %s", messages[0].Content)
	}
	if messages[1].Content != "Message 2" {
		t.Errorf("Expected second message to be 'Message 2', got %s", messages[1].Content)
	}
}

func TestConversationManager_Cleanup(t *testing.T) {
	cm := NewConversationManager(10)

	// Create conversation
	conversation := cm.Create("user1")

	// Set old timestamp
	cm.mu.Lock()
	cm.conversations[conversation.ID].UpdatedAt = time.Now().Add(-2 * time.Hour)
	cm.mu.Unlock()

	// Cleanup with 1 hour max age
	count := cm.Cleanup(1 * time.Hour)
	if count != 1 {
		t.Errorf("Expected 1 conversation to be cleaned up, got %d", count)
	}

	_, ok := cm.Get(conversation.ID)
	if ok {
		t.Error("Expected conversation to be cleaned up")
	}
}

func TestKnowledgeBase(t *testing.T) {
	tmpDir := createTempKnowledgeDir(t)
	defer os.RemoveAll(tmpDir)

	kb := NewKnowledgeBase(tmpDir)

	// Test Add
	article := &KnowledgeArticle{
		ID:       "test-article",
		Title:    "Test Article",
		Content:  "This is a test article",
		Category: "test",
		Tags:     []string{"test", "example"},
	}

	err := kb.Add(article)
	if err != nil {
		t.Fatalf("Failed to add article: %v", err)
	}

	// Test Get
	retrieved, ok := kb.Get("test-article")
	if !ok {
		t.Fatal("Expected article to be found")
	}

	if retrieved.Title != "Test Article" {
		t.Errorf("Expected title to be 'Test Article', got %s", retrieved.Title)
	}

	// Test Search
	results := kb.Search("test")
	if len(results) == 0 {
		t.Error("Expected search to find results")
	}

	// Test GetByCategory
	articles := kb.GetByCategory("test")
	if len(articles) == 0 {
		t.Error("Expected to find articles in test category")
	}

	// Test Count
	if kb.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", kb.Count())
	}

	// Test Delete
	err = kb.Delete("test-article")
	if err != nil {
		t.Fatalf("Failed to delete article: %v", err)
	}

	_, ok = kb.Get("test-article")
	if ok {
		t.Error("Expected article to be deleted")
	}
}

func TestKnowledgeBase_Search(t *testing.T) {
	tmpDir := createTempKnowledgeDir(t)
	defer os.RemoveAll(tmpDir)

	kb := NewKnowledgeBase(tmpDir)

	// Add test articles
	kb.Add(&KnowledgeArticle{
		ID:       "storage-1",
		Title:    "Storage Management",
		Content:  "How to manage storage pools",
		Category: "storage",
		Tags:     []string{"storage", "pool"},
	})

	kb.Add(&KnowledgeArticle{
		ID:       "network-1",
		Title:    "Network Configuration",
		Content:  "How to configure network settings",
		Category: "network",
		Tags:     []string{"network", "config"},
	})

	// Search for storage
	results := kb.Search("storage")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'storage', got %d", len(results))
	}

	// Search for network
	results = kb.Search("network")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'network', got %d", len(results))
	}

	// Search for non-existent
	results = kb.Search("nonexistent")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestKnowledgeBase_Categories(t *testing.T) {
	tmpDir := createTempKnowledgeDir(t)
	defer os.RemoveAll(tmpDir)

	kb := NewKnowledgeBase(tmpDir)

	kb.Add(&KnowledgeArticle{
		ID:       "1",
		Title:    "Article 1",
		Content:  "Content 1",
		Category: "storage",
		Tags:     []string{},
	})

	kb.Add(&KnowledgeArticle{
		ID:       "2",
		Title:    "Article 2",
		Content:  "Content 2",
		Category: "network",
		Tags:     []string{},
	})

	categories := kb.GetCategories()
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}

	// Check categories exist
	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}

	if !categoryMap["storage"] {
		t.Error("Expected storage category to exist")
	}
	if !categoryMap["network"] {
		t.Error("Expected network category to exist")
	}
}

func TestLLMClient(t *testing.T) {
	config := LLMConfig{
		BaseURL:     "http://localhost:11434",
		Model:       "llama2",
		Timeout:     30,
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	client := NewLLMClient(config)
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.config.Model != "llama2" {
		t.Errorf("Expected model to be llama2, got %s", client.config.Model)
	}
}

func TestModels(t *testing.T) {
	// Test QueryRequest
	req := QueryRequest{
		Query:  "test query",
		UserID: "user1",
	}

	if req.Query != "test query" {
		t.Errorf("Expected query to be 'test query', got %s", req.Query)
	}

	// Test QueryResponse
	resp := QueryResponse{
		Answer:         "test answer",
		ConversationID: "conv1",
		Sources:        []string{"source1"},
		Suggestions:    []string{"suggestion1"},
	}

	if resp.Answer != "test answer" {
		t.Errorf("Expected answer to be 'test answer', got %s", resp.Answer)
	}

	// Test DiagnosticResult
	result := DiagnosticResult{
		Issue:     "test issue",
		Severity:  "high",
		Causes:    []string{"cause1"},
		Solutions: []string{"solution1"},
	}

	if result.Severity != "high" {
		t.Errorf("Expected severity to be 'high', got %s", result.Severity)
	}

	// Test Recommendation
	rec := Recommendation{
		ID:          "rec1",
		Title:       "Test Recommendation",
		Description: "Test description",
		Category:    "test",
		Relevance:   0.8,
		Action:      "test action",
	}

	if rec.Relevance != 0.8 {
		t.Errorf("Expected relevance to be 0.8, got %f", rec.Relevance)
	}

	// Test HealthCheck
	health := HealthCheck{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	if health.Status != "ok" {
		t.Errorf("Expected status to be 'ok', got %s", health.Status)
	}
}

func TestAdvisor_AnalyzeInterests(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	// Create conversations with different interests
	conversations := []*Conversation{
		{
			ID:     "conv1",
			UserID: "user1",
			Messages: []Message{
				{Role: "user", Content: "How do I manage storage?"},
				{Role: "assistant", Content: "You can manage storage..."},
				{Role: "user", Content: "Tell me about RAID"},
			},
		},
		{
			ID:     "conv2",
			UserID: "user1",
			Messages: []Message{
				{Role: "user", Content: "How to set up VPN?"},
				{Role: "assistant", Content: "You can set up VPN..."},
			},
		},
	}

	interests := advisor.analyzeInterests(conversations)

	if interests["storage"] != 2 {
		t.Errorf("Expected storage interest to be 2, got %d", interests["storage"])
	}

	if interests["network"] != 1 {
		t.Errorf("Expected network interest to be 1, got %d", interests["network"])
	}
}

func TestAdvisor_GenerateRecommendations(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	// Test with interests
	interests := map[string]int{
		"storage": 5,
		"network": 3,
	}

	recommendations := advisor.generateRecommendations(interests)
	if len(recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}

	// Test with no interests
	emptyInterests := map[string]int{}
	recommendations = advisor.generateRecommendations(emptyInterests)
	if len(recommendations) == 0 {
		t.Error("Expected default recommendations to be generated")
	}

	if recommendations[0].ID != "rec-getting-started" {
		t.Errorf("Expected first recommendation to be 'rec-getting-started', got %s", recommendations[0].ID)
	}
}

func TestAdvisor_BuildSystemPrompt(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	// Test without knowledge context
	prompt := advisor.buildSystemPrompt("")
	if prompt == "" {
		t.Error("Expected system prompt to be generated")
	}

	// Test with knowledge context
	knowledgeContext := "## Test Article\nThis is test content\n"
	prompt = advisor.buildSystemPrompt(knowledgeContext)
	if prompt == "" {
		t.Error("Expected system prompt with knowledge context to be generated")
	}

	if !contains(prompt, "Test Article") {
		t.Error("Expected prompt to contain knowledge context")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

func TestAPI_Creation(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)
	api := NewAPI(advisor)

	if api == nil {
		t.Fatal("Expected API to be created")
	}

	if api.router == nil {
		t.Fatal("Expected router to be initialized")
	}
}

func TestConversationManager_GetOrCreate(t *testing.T) {
	cm := NewConversationManager(10)

	// Test with empty conversation ID
	conversation := cm.GetOrCreate("", "user1")
	if conversation == nil {
		t.Fatal("Expected conversation to be created")
	}

	if conversation.UserID != "user1" {
		t.Errorf("Expected user ID to be user1, got %s", conversation.UserID)
	}

	// Test with existing conversation ID
	existing := cm.Create("user2")
	retrieved := cm.GetOrCreate(existing.ID, "user2")
	if retrieved.ID != existing.ID {
		t.Errorf("Expected conversation ID to be %s, got %s", existing.ID, retrieved.ID)
	}

	// Test with non-existent conversation ID
	nonExistent := cm.GetOrCreate("non-existent", "user3")
	if nonExistent == nil {
		t.Fatal("Expected new conversation to be created")
	}

	if nonExistent.UserID != "user3" {
		t.Errorf("Expected user ID to be user3, got %s", nonExistent.UserID)
	}
}

func TestConversationManager_ToLLMMessages(t *testing.T) {
	cm := NewConversationManager(10)

	conversation := cm.Create("user1")
	cm.AddMessage(conversation.ID, "user", "Hello")
	cm.AddMessage(conversation.ID, "assistant", "Hi there!")

	// Test without system prompt
	messages, err := cm.ToLLMMessages(conversation.ID, "")
	if err != nil {
		t.Fatalf("Failed to convert to LLM messages: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Test with system prompt
	messages, err = cm.ToLLMMessages(conversation.ID, "You are a helpful assistant")
	if err != nil {
		t.Fatalf("Failed to convert to LLM messages: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages (with system prompt), got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected first message role to be 'system', got %s", messages[0].Role)
	}

	// Test with non-existent conversation
	_, err = cm.ToLLMMessages("non-existent", "")
	if err == nil {
		t.Error("Expected error for non-existent conversation")
	}
}

func TestAdvisor_GetStatus(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)
	status := advisor.GetStatus()

	if status == nil {
		t.Fatal("Expected status to be returned")
	}

	if status.Status != "running" {
		t.Errorf("Expected status to be 'running', got %s", status.Status)
	}
}

func TestAdvisor_HealthCheck(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)
	health := advisor.HealthCheck()

	if health == nil {
		t.Fatal("Expected health check to be returned")
	}

	if health.Status != "ok" {
		t.Errorf("Expected health status to be 'ok', got %s", health.Status)
	}

	if health.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got %s", health.Version)
	}
}

func TestBuildKnowledgeContext(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	// Test with empty articles
	context := advisor.buildKnowledgeContext([]*KnowledgeArticle{})
	if context != "" {
		t.Errorf("Expected empty context for empty articles, got: %s", context)
	}

	// Test with articles
	articles := []*KnowledgeArticle{
		{
			ID:      "1",
			Title:   "Test Article 1",
			Content: "Content 1",
		},
		{
			ID:      "2",
			Title:   "Test Article 2",
			Content: "Content 2",
		},
	}

	context = advisor.buildKnowledgeContext(articles)
	if context == "" {
		t.Error("Expected context to be generated")
	}

	if !contains(context, "Test Article 1") {
		t.Error("Expected context to contain article titles")
	}
}

func TestGenerateSuggestions(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	// Test with storage articles
	articles := []*KnowledgeArticle{
		{Category: "storage"},
	}

	suggestions := advisor.generateSuggestions("storage query", articles)
	if len(suggestions) == 0 {
		t.Error("Expected suggestions to be generated")
	}

	// Verify suggestions are limited to 3
	if len(suggestions) > 3 {
		t.Errorf("Expected max 3 suggestions, got %d", len(suggestions))
	}
}

func TestBuildDiagnosticPrompt(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	req := &DiagnosticRequest{
		UserID:   "user1",
		Symptoms: []string{"slow performance", "high CPU usage"},
		SystemInfo: "OS: Linux\nCPU: 4 cores\nRAM: 8GB",
	}

	prompt := advisor.buildDiagnosticPrompt(req)
	if prompt == "" {
		t.Error("Expected diagnostic prompt to be generated")
	}

	if !contains(prompt, "slow performance") {
		t.Error("Expected prompt to contain symptoms")
	}

	if !contains(prompt, "OS: Linux") {
		t.Error("Expected prompt to contain system info")
	}
}

func TestParseDiagnosticResponse(t *testing.T) {
	config := AdvisorConfig{
		LLM: LLMConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Timeout:     30,
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		KnowledgePath:      os.TempDir(),
		MaxMessages:        50,
		CleanupInterval:    60,
		MaxConversationAge: 24,
	}

	advisor := NewAdvisor(config)

	// Test with critical symptoms
	req := &DiagnosticRequest{
		UserID:   "user1",
		Symptoms: []string{"critical failure"},
	}

	response := "The system has experienced a critical failure..."
	result := advisor.parseDiagnosticResponse(response, req)

	if result.Severity != "high" {
		t.Errorf("Expected severity to be 'high', got %s", result.Severity)
	}

	// Test with performance symptoms
	req = &DiagnosticRequest{
		UserID:   "user1",
		Symptoms: []string{"slow performance"},
	}

	response = "The system is experiencing slow performance..."
	result = advisor.parseDiagnosticResponse(response, req)

	if result.Severity != "medium" {
		t.Errorf("Expected severity to be 'medium', got %s", result.Severity)
	}
}
