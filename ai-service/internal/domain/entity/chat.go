// Package entity defines domain entities for AI Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage represents a message in a chat conversation.
type ChatMessage struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" bson:"_id"`

	// SceneID is the scene this chat belongs to.
	SceneID string `json:"scene_id" bson:"scene_id"`

	// BranchID is the branch this chat belongs to.
	BranchID string `json:"branch_id" bson:"branch_id"`

	// ContextID groups messages in a conversation.
	ContextID string `json:"context_id" bson:"context_id"`

	// Role is "user" or "assistant".
	Role string `json:"role" bson:"role"`

	// Content is the message text.
	Content string `json:"content" bson:"content"`

	// Actions are suggested actions (for assistant messages).
	Actions []SuggestedAction `json:"actions,omitempty" bson:"actions,omitempty"`

	// TokenUsage for assistant messages.
	TokenUsage *TokenUsage `json:"token_usage,omitempty" bson:"token_usage,omitempty"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// SuggestedAction represents an action suggested by AI.
type SuggestedAction struct {
	// ID is the action identifier.
	ID string `json:"id" bson:"id"`

	// Type is the action type (e.g., "DEMOLISH_WALL", "ADD_FURNITURE").
	Type string `json:"type" bson:"type"`

	// Description for the user.
	Description string `json:"description" bson:"description"`

	// Params are parameters for the action.
	Params map[string]string `json:"params" bson:"params"`

	// Confidence in the recommendation (0-1).
	Confidence float64 `json:"confidence" bson:"confidence"`

	// RequiresConfirmation if user must approve.
	RequiresConfirmation bool `json:"requires_confirmation" bson:"requires_confirmation"`
}

// TokenUsage represents LLM token usage.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens" bson:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens" bson:"completion_tokens"`
	TotalTokens      int `json:"total_tokens" bson:"total_tokens"`
}

// NewChatMessage creates a new chat message.
func NewChatMessage(sceneID, branchID, contextID, role, content string) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New(),
		SceneID:   sceneID,
		BranchID:  branchID,
		ContextID: contextID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
}

// WithActions adds suggested actions.
func (m *ChatMessage) WithActions(actions []SuggestedAction) *ChatMessage {
	m.Actions = actions
	return m
}

// WithTokenUsage adds token usage.
func (m *ChatMessage) WithTokenUsage(usage *TokenUsage) *ChatMessage {
	m.TokenUsage = usage
	return m
}

// AIContext represents the AI context for a scene/branch.
type AIContext struct {
	// ID is the context identifier.
	ID string `json:"id" bson:"_id"`

	// SceneID is the scene this context belongs to.
	SceneID string `json:"scene_id" bson:"scene_id"`

	// BranchID is the branch this context belongs to.
	BranchID string `json:"branch_id" bson:"branch_id"`

	// SceneSummary is a text summary of the scene for AI.
	SceneSummary string `json:"scene_summary" bson:"scene_summary"`

	// Constraints are active constraints.
	Constraints []string `json:"constraints" bson:"constraints"`

	// ContextSize is the size in tokens.
	ContextSize int `json:"context_size" bson:"context_size"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// NewAIContext creates a new AI context.
func NewAIContext(sceneID, branchID string) *AIContext {
	return &AIContext{
		ID:        uuid.New().String(),
		SceneID:   sceneID,
		BranchID:  branchID,
		UpdatedAt: time.Now().UTC(),
	}
}
