// Package entity provides tests for domain entities.
package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNewChatMessage verifies chat message creation.
func TestNewChatMessage(t *testing.T) {
	t.Parallel()

	sceneID := "scene-123"
	branchID := "branch-456"
	contextID := "context-789"
	role := "user"
	content := "Hello, AI assistant!"

	msg := NewChatMessage(sceneID, branchID, contextID, role, content)

	if msg == nil {
		t.Fatal("expected non-nil message")
	}

	if msg.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if msg.SceneID != sceneID {
		t.Errorf("expected scene_id %s, got %s", sceneID, msg.SceneID)
	}

	if msg.BranchID != branchID {
		t.Errorf("expected branch_id %s, got %s", branchID, msg.BranchID)
	}

	if msg.ContextID != contextID {
		t.Errorf("expected context_id %s, got %s", contextID, msg.ContextID)
	}

	if msg.Role != role {
		t.Errorf("expected role %s, got %s", role, msg.Role)
	}

	if msg.Content != content {
		t.Errorf("expected content %s, got %s", content, msg.Content)
	}

	if msg.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

// TestChatMessage_WithActions verifies action assignment.
func TestChatMessage_WithActions(t *testing.T) {
	t.Parallel()

	msg := NewChatMessage("scene", "branch", "ctx", "assistant", "Here's what I suggest")

	actions := []SuggestedAction{
		{
			ID:                   "action-1",
			Type:                 "DEMOLISH_WALL",
			Description:          "Remove the partition wall",
			Confidence:           0.9,
			RequiresConfirmation: true,
		},
		{
			ID:                   "action-2",
			Type:                 "ADD_FURNITURE",
			Description:          "Add a sofa",
			Confidence:           0.8,
			RequiresConfirmation: false,
		},
	}

	msg.WithActions(actions)

	if len(msg.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(msg.Actions))
	}

	if msg.Actions[0].Type != "DEMOLISH_WALL" {
		t.Errorf("expected first action type DEMOLISH_WALL, got %s", msg.Actions[0].Type)
	}
}

// TestChatMessage_WithTokenUsage verifies token usage assignment.
func TestChatMessage_WithTokenUsage(t *testing.T) {
	t.Parallel()

	msg := NewChatMessage("scene", "branch", "ctx", "assistant", "Response")

	usage := &TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	msg.WithTokenUsage(usage)

	if msg.TokenUsage == nil {
		t.Fatal("expected non-nil token usage")
	}

	if msg.TokenUsage.PromptTokens != 100 {
		t.Errorf("expected prompt tokens 100, got %d", msg.TokenUsage.PromptTokens)
	}

	if msg.TokenUsage.CompletionTokens != 50 {
		t.Errorf("expected completion tokens 50, got %d", msg.TokenUsage.CompletionTokens)
	}

	if msg.TokenUsage.TotalTokens != 150 {
		t.Errorf("expected total tokens 150, got %d", msg.TokenUsage.TotalTokens)
	}
}

// TestNewAIContext verifies AI context creation.
func TestNewAIContext(t *testing.T) {
	t.Parallel()

	sceneID := "scene-123"
	branchID := "branch-456"

	ctx := NewAIContext(sceneID, branchID)

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	if ctx.ID == "" {
		t.Error("expected non-empty ID")
	}

	if ctx.SceneID != sceneID {
		t.Errorf("expected scene_id %s, got %s", sceneID, ctx.SceneID)
	}

	if ctx.BranchID != branchID {
		t.Errorf("expected branch_id %s, got %s", branchID, ctx.BranchID)
	}

	if ctx.UpdatedAt.IsZero() {
		t.Error("expected non-zero updated_at")
	}
}

// TestSuggestedAction verifies suggested action structure.
func TestSuggestedAction(t *testing.T) {
	t.Parallel()

	action := SuggestedAction{
		ID:          "action-id",
		Type:        "CHANGE_ROOM_TYPE",
		Description: "Convert bedroom to office",
		Params: map[string]string{
			"room_id":  "room-123",
			"new_type": "OFFICE",
		},
		Confidence:           0.85,
		RequiresConfirmation: true,
	}

	if action.ID != "action-id" {
		t.Errorf("expected ID action-id, got %s", action.ID)
	}

	if action.Type != "CHANGE_ROOM_TYPE" {
		t.Errorf("expected type CHANGE_ROOM_TYPE, got %s", action.Type)
	}

	if action.Params["room_id"] != "room-123" {
		t.Errorf("expected param room_id=room-123, got %s", action.Params["room_id"])
	}

	if action.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", action.Confidence)
	}

	if !action.RequiresConfirmation {
		t.Error("expected requires_confirmation to be true")
	}
}

// TestTokenUsage verifies token usage calculation.
func TestTokenUsage(t *testing.T) {
	t.Parallel()

	usage := &TokenUsage{
		PromptTokens:     250,
		CompletionTokens: 100,
		TotalTokens:      350,
	}

	// Verify total equals sum
	if usage.TotalTokens != usage.PromptTokens+usage.CompletionTokens {
		t.Errorf("total tokens (%d) should equal prompt (%d) + completion (%d)",
			usage.TotalTokens, usage.PromptTokens, usage.CompletionTokens)
	}
}

// BenchmarkNewChatMessage benchmarks chat message creation.
func BenchmarkNewChatMessage(b *testing.B) {
	sceneID := "scene-123"
	branchID := "branch-456"
	contextID := "context-789"
	role := "user"
	content := "This is a test message for benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewChatMessage(sceneID, branchID, contextID, role, content)
	}
}

// BenchmarkWithActions benchmarks action assignment.
func BenchmarkWithActions(b *testing.B) {
	msg := NewChatMessage("scene", "branch", "ctx", "assistant", "Response")
	actions := []SuggestedAction{
		{ID: "1", Type: "TEST", Confidence: 0.9},
		{ID: "2", Type: "TEST2", Confidence: 0.8},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.WithActions(actions)
	}
}

// TestChatMessageTimestamp verifies timestamps are set correctly.
func TestChatMessageTimestamp(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC()
	msg := NewChatMessage("scene", "branch", "ctx", "user", "test")
	after := time.Now().UTC()

	if msg.CreatedAt.Before(before) {
		t.Error("created_at should not be before test start")
	}

	if msg.CreatedAt.After(after) {
		t.Error("created_at should not be after test end")
	}
}
