// =============================================================================
// Package grpc provides gRPC handlers and clients for AI Service.
// =============================================================================
// SceneClient provides integration with Scene Service for getting scene data.
// Used by AI services to understand the current layout context.
// =============================================================================
package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	scenepb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// SceneClient wraps Scene Service gRPC client for AI context integration.
// It provides methods to fetch scene data formatted for AI prompts.
type SceneClient struct {
	client scenepb.SceneServiceClient
	conn   *grpc.ClientConn
	log    *logger.Logger

	// Cache for scene context (to reduce gRPC calls)
	cache    map[string]*cachedContext
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
}

// cachedContext holds cached scene context with expiration.
type cachedContext struct {
	context   string
	expiresAt time.Time
}

// SceneContextData represents scene data formatted for AI.
type SceneContextData struct {
	SceneID    string          `json:"scene_id"`
	Name       string          `json:"name"`
	TotalArea  float64         `json:"total_area"`
	Dimensions Dimensions2D    `json:"dimensions"`
	Walls      []WallInfo      `json:"walls"`
	Rooms      []RoomInfo      `json:"rooms"`
	Openings   []OpeningInfo   `json:"openings"`
	Furniture  []FurnitureInfo `json:"furniture,omitempty"`
}

// Dimensions2D represents 2D dimensions.
type Dimensions2D struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// WallInfo represents wall information for AI context.
type WallInfo struct {
	ID            string  `json:"id"`
	IsLoadBearing bool    `json:"is_load_bearing"`
	Thickness     float64 `json:"thickness"`
	Material      string  `json:"material,omitempty"`
}

// RoomInfo represents room information for AI context.
type RoomInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Area      float64 `json:"area"`
	IsWetZone bool    `json:"is_wet_zone"`
}

// OpeningInfo represents opening (door/window) information for AI context.
type OpeningInfo struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"` // door, window, arch
	Width  float64 `json:"width"`
	WallID string  `json:"wall_id,omitempty"`
}

// FurnitureInfo represents furniture information for AI context.
type FurnitureInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// NewSceneClient creates a new Scene Service gRPC client.
// addr should be in format "host:port" (e.g., "scene-service:50051").
func NewSceneClient(addr string, log *logger.Logger) (*SceneClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("scene service address is required")
	}

	// Create gRPC connection with retry options
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to scene service at %s: %w", addr, err)
	}

	log.Info("connected to Scene Service",
		logger.String("address", addr),
	)

	return &SceneClient{
		client:   scenepb.NewSceneServiceClient(conn),
		conn:     conn,
		log:      log,
		cache:    make(map[string]*cachedContext),
		cacheTTL: 5 * time.Minute, // Cache scene context for 5 minutes
	}, nil
}

// Close closes the gRPC connection.
func (c *SceneClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetSceneContext returns scene data formatted for AI context.
// The result is a human-readable JSON description suitable for AI prompts.
// Results are cached for cacheTTL duration.
func (c *SceneClient) GetSceneContext(ctx context.Context, sceneID string) (string, error) {
	if sceneID == "" {
		return "", fmt.Errorf("scene_id is required")
	}

	// Check cache first
	c.cacheMu.RLock()
	if cached, ok := c.cache[sceneID]; ok && time.Now().Before(cached.expiresAt) {
		c.cacheMu.RUnlock()
		c.log.Debug("returning cached scene context",
			logger.String("scene_id", sceneID),
		)
		return cached.context, nil
	}
	c.cacheMu.RUnlock()

	// Fetch scene from Scene Service
	scene, err := c.client.GetScene(ctx, &scenepb.GetSceneRequest{Id: sceneID})
	if err != nil {
		c.log.Warn("failed to get scene from Scene Service",
			logger.Err(err),
			logger.String("scene_id", sceneID),
		)
		return "", fmt.Errorf("failed to get scene: %w", err)
	}

	// Fetch elements for this scene
	// Note: ListElementsRequest uses branch_id, so we use main_branch_id
	branchID := scene.Scene.MainBranchId
	elements, err := c.client.ListElements(ctx, &scenepb.ListElementsRequest{
		BranchId: branchID,
		Limit:    1000, // Get all elements
	})
	if err != nil {
		c.log.Warn("failed to list elements from Scene Service",
			logger.Err(err),
			logger.String("scene_id", sceneID),
			logger.String("branch_id", branchID),
		)
		// Continue without elements - we still have scene info
		elements = &scenepb.ListElementsResponse{}
	}

	// Build context data structure
	contextData := SceneContextData{
		SceneID:   sceneID,
		Name:      scene.Scene.Name,
		Walls:     make([]WallInfo, 0),
		Rooms:     make([]RoomInfo, 0),
		Openings:  make([]OpeningInfo, 0),
		Furniture: make([]FurnitureInfo, 0),
	}

	// Parse dimensions if available
	if scene.Scene.Dimensions != nil {
		contextData.Dimensions = Dimensions2D{
			Width:  float64(scene.Scene.Dimensions.Width),
			Height: float64(scene.Scene.Dimensions.Height),
		}
		// Estimate total area from dimensions (approximate)
		contextData.TotalArea = contextData.Dimensions.Width * contextData.Dimensions.Height
	}

	// Categorize elements by type
	for _, el := range elements.Elements {
		switch el.Type {
		case scenepb.ElementType_ELEMENT_TYPE_WALL:
			contextData.Walls = append(contextData.Walls, WallInfo{
				ID:            el.Id,
				IsLoadBearing: false, // Default, would need element properties
				Thickness:     0.2,   // Default thickness
			})
		case scenepb.ElementType_ELEMENT_TYPE_ROOM:
			contextData.Rooms = append(contextData.Rooms, RoomInfo{
				ID:   el.Id,
				Name: el.Name,
				Type: "unknown", // Would need element properties
			})
		case scenepb.ElementType_ELEMENT_TYPE_DOOR:
			contextData.Openings = append(contextData.Openings, OpeningInfo{
				ID:   el.Id,
				Type: "door",
			})
		case scenepb.ElementType_ELEMENT_TYPE_WINDOW:
			contextData.Openings = append(contextData.Openings, OpeningInfo{
				ID:   el.Id,
				Type: "window",
			})
		case scenepb.ElementType_ELEMENT_TYPE_FURNITURE:
			contextData.Furniture = append(contextData.Furniture, FurnitureInfo{
				ID:   el.Id,
				Name: el.Name,
			})
		}
	}

	// Format as JSON for AI prompt
	jsonBytes, err := json.MarshalIndent(contextData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal scene context: %w", err)
	}

	contextString := fmt.Sprintf("Текущая планировка:\n```json\n%s\n```", string(jsonBytes))

	// Cache the result
	c.cacheMu.Lock()
	c.cache[sceneID] = &cachedContext{
		context:   contextString,
		expiresAt: time.Now().Add(c.cacheTTL),
	}
	c.cacheMu.Unlock()

	c.log.Debug("fetched scene context",
		logger.String("scene_id", sceneID),
		logger.Int("walls", len(contextData.Walls)),
		logger.Int("rooms", len(contextData.Rooms)),
	)

	return contextString, nil
}

// InvalidateCache removes a scene from the cache.
// Call this when the scene is modified.
func (c *SceneClient) InvalidateCache(sceneID string) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	delete(c.cache, sceneID)
}

// ClearCache removes all entries from the cache.
func (c *SceneClient) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]*cachedContext)
}

// GetSceneContextForBranch returns scene context for a specific branch.
// Useful when comparing different layout variants.
func (c *SceneClient) GetSceneContextForBranch(ctx context.Context, sceneID, branchID string) (string, error) {
	if sceneID == "" || branchID == "" {
		return "", fmt.Errorf("scene_id and branch_id are required")
	}

	// Fetch scene
	scene, err := c.client.GetScene(ctx, &scenepb.GetSceneRequest{Id: sceneID})
	if err != nil {
		return "", fmt.Errorf("failed to get scene: %w", err)
	}

	// Fetch elements for specific branch
	elements, err := c.client.ListElements(ctx, &scenepb.ListElementsRequest{
		BranchId: branchID,
		Limit:    1000,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list elements: %w", err)
	}

	// Build context (same as GetSceneContext but for specific branch)
	contextData := SceneContextData{
		SceneID:   sceneID,
		Name:      scene.Scene.Name + " (ветка: " + branchID + ")",
		Walls:     make([]WallInfo, 0),
		Rooms:     make([]RoomInfo, 0),
		Openings:  make([]OpeningInfo, 0),
		Furniture: make([]FurnitureInfo, 0),
	}

	if scene.Scene.Dimensions != nil {
		contextData.Dimensions = Dimensions2D{
			Width:  float64(scene.Scene.Dimensions.Width),
			Height: float64(scene.Scene.Dimensions.Height),
		}
		contextData.TotalArea = contextData.Dimensions.Width * contextData.Dimensions.Height
	}

	for _, el := range elements.Elements {
		switch el.Type {
		case scenepb.ElementType_ELEMENT_TYPE_WALL:
			contextData.Walls = append(contextData.Walls, WallInfo{ID: el.Id})
		case scenepb.ElementType_ELEMENT_TYPE_ROOM:
			contextData.Rooms = append(contextData.Rooms, RoomInfo{ID: el.Id, Name: el.Name})
		case scenepb.ElementType_ELEMENT_TYPE_DOOR:
			contextData.Openings = append(contextData.Openings, OpeningInfo{ID: el.Id, Type: "door"})
		case scenepb.ElementType_ELEMENT_TYPE_WINDOW:
			contextData.Openings = append(contextData.Openings, OpeningInfo{ID: el.Id, Type: "window"})
		case scenepb.ElementType_ELEMENT_TYPE_FURNITURE:
			contextData.Furniture = append(contextData.Furniture, FurnitureInfo{ID: el.Id, Name: el.Name})
		}
	}

	jsonBytes, err := json.MarshalIndent(contextData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal scene context: %w", err)
	}

	return fmt.Sprintf("Планировка (ветка %s):\n```json\n%s\n```", branchID, string(jsonBytes)), nil
}
