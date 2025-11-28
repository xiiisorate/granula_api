// Package mongodb provides MongoDB implementations of repositories.
package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/ai-service/internal/domain/entity"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChatRepository handles chat message persistence.
type ChatRepository struct {
	collection *mongo.Collection
}

// NewChatRepository creates a new ChatRepository.
func NewChatRepository(db *mongo.Database) *ChatRepository {
	collection := db.Collection("chat_messages")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "scene_id", Value: 1},
				{Key: "branch_id", Value: 1},
				{Key: "context_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "scene_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &ChatRepository{collection: collection}
}

// Save saves a chat message.
func (r *ChatRepository) Save(ctx context.Context, msg *entity.ChatMessage) error {
	_, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return apperrors.Internal("failed to save chat message").WithCause(err)
	}
	return nil
}

// GetByID retrieves a message by ID.
func (r *ChatRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ChatMessage, error) {
	var msg entity.ChatMessage
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&msg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("chat_message", id.String())
		}
		return nil, apperrors.Internal("failed to get chat message").WithCause(err)
	}
	return &msg, nil
}

// GetHistory retrieves chat history for a scene/branch.
func (r *ChatRepository) GetHistory(ctx context.Context, sceneID, branchID string, opts GetHistoryOptions) ([]*entity.ChatMessage, bool, error) {
	filter := bson.M{
		"scene_id":  sceneID,
		"branch_id": branchID,
	}

	if opts.ContextID != "" {
		filter["context_id"] = opts.ContextID
	}

	// Add cursor filter
	if opts.Cursor != "" {
		cursorID, err := uuid.Parse(opts.Cursor)
		if err == nil {
			// Get cursor message to find its timestamp
			var cursorMsg entity.ChatMessage
			err := r.collection.FindOne(ctx, bson.M{"_id": cursorID}).Decode(&cursorMsg)
			if err == nil {
				filter["created_at"] = bson.M{"$lt": cursorMsg.CreatedAt}
			}
		}
	}

	// Set limit + 1 to check if there are more
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit + 1))

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, false, apperrors.Internal("failed to get chat history").WithCause(err)
	}
	defer cursor.Close(ctx)

	var messages []*entity.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, false, apperrors.Internal("failed to decode chat history").WithCause(err)
	}

	// Check if there are more messages
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, hasMore, nil
}

// GetHistoryOptions for fetching chat history.
type GetHistoryOptions struct {
	ContextID string
	Limit     int
	Cursor    string
}

// DeleteHistory deletes chat history for a scene/branch.
func (r *ChatRepository) DeleteHistory(ctx context.Context, sceneID, branchID, contextID string) (int64, error) {
	filter := bson.M{
		"scene_id":  sceneID,
		"branch_id": branchID,
	}

	if contextID != "" {
		filter["context_id"] = contextID
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, apperrors.Internal("failed to delete chat history").WithCause(err)
	}

	return result.DeletedCount, nil
}

// GetRecentMessages gets the most recent messages for context building.
func (r *ChatRepository) GetRecentMessages(ctx context.Context, sceneID, branchID, contextID string, limit int) ([]*entity.ChatMessage, error) {
	filter := bson.M{
		"scene_id":  sceneID,
		"branch_id": branchID,
	}

	if contextID != "" {
		filter["context_id"] = contextID
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, apperrors.Internal("failed to get recent messages").WithCause(err)
	}
	defer cursor.Close(ctx)

	var messages []*entity.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, apperrors.Internal("failed to decode messages").WithCause(err)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountMessages counts messages for a scene/branch.
func (r *ChatRepository) CountMessages(ctx context.Context, sceneID, branchID string) (int64, error) {
	filter := bson.M{
		"scene_id":  sceneID,
		"branch_id": branchID,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, apperrors.Internal("failed to count messages").WithCause(err)
	}

	return count, nil
}
