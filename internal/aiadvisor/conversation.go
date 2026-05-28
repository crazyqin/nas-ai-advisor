package aiadvisor

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConversationManager manages conversation sessions
type ConversationManager struct {
	mu            sync.RWMutex
	conversations map[string]*Conversation
	maxMessages   int
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(maxMessages int) *ConversationManager {
	return &ConversationManager{
		conversations: make(map[string]*Conversation),
		maxMessages:   maxMessages,
	}
}

// Create creates a new conversation
func (cm *ConversationManager) Create(userID string) *Conversation {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conversation := &Conversation{
		ID:        uuid.New().String(),
		UserID:    userID,
		Messages:  make([]Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cm.conversations[conversation.ID] = conversation
	return conversation
}

// Get retrieves a conversation by ID
func (cm *ConversationManager) Get(id string) (*Conversation, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conversation, ok := cm.conversations[id]
	return conversation, ok
}

// GetOrCreate retrieves or creates a conversation
func (cm *ConversationManager) GetOrCreate(conversationID, userID string) *Conversation {
	if conversationID != "" {
		if conversation, ok := cm.Get(conversationID); ok {
			return conversation
		}
	}

	return cm.Create(userID)
}

// AddMessage adds a message to a conversation
func (cm *ConversationManager) AddMessage(conversationID, role, content string) (*Message, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conversation, ok := cm.conversations[conversationID]
	if !ok {
		return nil, fmt.Errorf("conversation %s not found", conversationID)
	}

	message := Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	conversation.Messages = append(conversation.Messages, message)
	conversation.UpdatedAt = time.Now()

	// Trim messages if exceeding limit
	if len(conversation.Messages) > cm.maxMessages {
		conversation.Messages = conversation.Messages[len(conversation.Messages)-cm.maxMessages:]
	}

	return &message, nil
}

// GetMessages retrieves all messages in a conversation
func (cm *ConversationManager) GetMessages(conversationID string) ([]Message, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conversation, ok := cm.conversations[conversationID]
	if !ok {
		return nil, fmt.Errorf("conversation %s not found", conversationID)
	}

	return conversation.Messages, nil
}

// GetUserConversations retrieves all conversations for a user
func (cm *ConversationManager) GetUserConversations(userID string) []*Conversation {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var conversations []*Conversation
	for _, conversation := range cm.conversations {
		if conversation.UserID == userID {
			conversations = append(conversations, conversation)
		}
	}

	return conversations
}

// Delete deletes a conversation
func (cm *ConversationManager) Delete(conversationID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.conversations[conversationID]; !ok {
		return fmt.Errorf("conversation %s not found", conversationID)
	}

	delete(cm.conversations, conversationID)
	return nil
}

// Cleanup removes old conversations
func (cm *ConversationManager) Cleanup(maxAge time.Duration) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	count := 0

	for id, conversation := range cm.conversations {
		if conversation.UpdatedAt.Before(cutoff) {
			delete(cm.conversations, id)
			count++
		}
	}

	return count
}

// Count returns the number of active conversations
func (cm *ConversationManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.conversations)
}

// ToLLMMessages converts conversation messages to LLM messages
func (cm *ConversationManager) ToLLMMessages(conversationID string, systemPrompt string) ([]LLMMessage, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conversation, ok := cm.conversations[conversationID]
	if !ok {
		return nil, fmt.Errorf("conversation %s not found", conversationID)
	}

	messages := make([]LLMMessage, 0, len(conversation.Messages)+1)

	// Add system prompt if provided
	if systemPrompt != "" {
		messages = append(messages, LLMMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range conversation.Messages {
		messages = append(messages, LLMMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return messages, nil
}
