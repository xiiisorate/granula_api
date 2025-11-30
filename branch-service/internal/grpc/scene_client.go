// =============================================================================
// Package grpc provides gRPC handlers and clients for Branch Service.
// =============================================================================
// SceneClient provides integration with Scene Service for managing scene elements
// when creating or merging branches.
//
// Usage:
//
//	client, err := NewSceneClient("scene-service:50055", log)
//	if err != nil {
//	    log.Warn("scene service unavailable", logger.Err(err))
//	}
//	// Use client in BranchService for element operations
//
// =============================================================================
package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	scenepb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// sceneConnTimeout is the timeout for connecting to scene service.
	sceneConnTimeout = 5 * time.Second

	// sceneCallTimeout is the default timeout for scene RPC calls.
	sceneCallTimeout = 30 * time.Second

	// maxElementsPerFetch is the maximum number of elements to fetch in one call.
	maxElementsPerFetch = 10000
)

// =============================================================================
// SceneClient
// =============================================================================

// SceneClient wraps Scene Service gRPC client.
// It provides methods for managing scene elements when working with branches.
//
// Thread Safety: Safe for concurrent use.
type SceneClient struct {
	client scenepb.SceneServiceClient
	conn   *grpc.ClientConn
	log    *logger.Logger
}

// NewSceneClient creates a new Scene Service gRPC client.
//
// Parameters:
//   - addr: gRPC address in format "host:port" (e.g., "scene-service:50055")
//   - log: Logger instance for operational logging
//
// Returns:
//   - *SceneClient: Connected client ready for use
//   - error: Connection error if service is unavailable
func NewSceneClient(addr string, log *logger.Logger) (*SceneClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("scene service address is required")
	}

	// Create gRPC connection with options
	ctx, cancel := context.WithTimeout(context.Background(), sceneConnTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to scene service at %s: %w", addr, err)
	}

	log.Info("connected to Scene Service",
		logger.String("address", addr),
	)

	return &SceneClient{
		client: scenepb.NewSceneServiceClient(conn),
		conn:   conn,
		log:    log,
	}, nil
}

// Close closes the gRPC connection.
// Should be called when the service shuts down.
func (c *SceneClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// =============================================================================
// Element Operations
// =============================================================================

// GetElements returns all elements for a branch.
// Used when creating a new branch to copy elements from parent.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - branchID: UUID of the branch to get elements for
//
// Returns:
//   - []*scenepb.Element: List of elements in the branch
//   - error: nil on success, error if retrieval failed
func (c *SceneClient) GetElements(ctx context.Context, branchID uuid.UUID) ([]*scenepb.Element, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("scene client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, sceneCallTimeout)
	defer cancel()

	resp, err := c.client.ListElements(ctx, &scenepb.ListElementsRequest{
		BranchId: branchID.String(),
		Limit:    maxElementsPerFetch,
	})
	if err != nil {
		c.log.Warn("failed to list elements from Scene Service",
			logger.Err(err),
			logger.String("branch_id", branchID.String()),
		)
		return nil, fmt.Errorf("list elements: %w", err)
	}

	c.log.Debug("fetched elements from Scene Service",
		logger.String("branch_id", branchID.String()),
		logger.Int("count", len(resp.Elements)),
	)

	return resp.Elements, nil
}

// CopyElementsToBranch copies all elements from source branch to target branch.
// Used when creating a new branch from a parent branch.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sceneID: UUID of the scene containing both branches
//   - sourceBranchID: UUID of the source branch to copy from
//   - targetBranchID: UUID of the target branch to copy to
//
// Returns:
//   - int: Number of elements copied
//   - error: nil on success, error if copy failed
//
// Note: This creates new element instances in the target branch,
// preserving all properties from the source elements.
func (c *SceneClient) CopyElementsToBranch(ctx context.Context, sceneID, sourceBranchID, targetBranchID uuid.UUID) (int, error) {
	if c == nil || c.client == nil {
		return 0, fmt.Errorf("scene client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, sceneCallTimeout*2) // Extended timeout for batch operation
	defer cancel()

	// Fetch all elements from source branch
	elements, err := c.GetElements(ctx, sourceBranchID)
	if err != nil {
		return 0, fmt.Errorf("get source elements: %w", err)
	}

	if len(elements) == 0 {
		c.log.Debug("no elements to copy",
			logger.String("source_branch_id", sourceBranchID.String()),
			logger.String("target_branch_id", targetBranchID.String()),
		)
		return 0, nil
	}

	// Copy each element to the target branch
	copiedCount := 0
	for _, el := range elements {
		_, err := c.client.CreateElement(ctx, &scenepb.CreateElementRequest{
			SceneId:    sceneID.String(),
			BranchId:   targetBranchID.String(),
			Type:       el.Type,
			Name:       el.Name,
			Position:   el.Position,
			Rotation:   el.Rotation,
			Dimensions: el.Dimensions,
			ParentId:   el.ParentId,
		})

		if err != nil {
			c.log.Warn("failed to copy element to target branch",
				logger.Err(err),
				logger.String("element_id", el.Id),
				logger.String("target_branch_id", targetBranchID.String()),
			)
			// Continue with other elements despite error
			continue
		}

		copiedCount++
	}

	c.log.Info("elements copied to branch",
		logger.String("source_branch_id", sourceBranchID.String()),
		logger.String("target_branch_id", targetBranchID.String()),
		logger.Int("total_elements", len(elements)),
		logger.Int("copied_elements", copiedCount),
	)

	return copiedCount, nil
}

// GetElementsForMerge returns elements from both branches for merge comparison.
// Used when merging branches to detect conflicts and differences.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sourceBranchID: UUID of the source branch (being merged)
//   - targetBranchID: UUID of the target branch (merge destination)
//
// Returns:
//   - sourceElements: Elements in the source branch
//   - targetElements: Elements in the target branch
//   - error: nil on success, error if retrieval failed
func (c *SceneClient) GetElementsForMerge(ctx context.Context, sourceBranchID, targetBranchID uuid.UUID) (sourceElements, targetElements []*scenepb.Element, err error) {
	if c == nil || c.client == nil {
		return nil, nil, fmt.Errorf("scene client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, sceneCallTimeout)
	defer cancel()

	// Fetch source elements
	sourceElements, err = c.GetElements(ctx, sourceBranchID)
	if err != nil {
		return nil, nil, fmt.Errorf("get source elements: %w", err)
	}

	// Fetch target elements
	targetElements, err = c.GetElements(ctx, targetBranchID)
	if err != nil {
		return nil, nil, fmt.Errorf("get target elements: %w", err)
	}

	c.log.Debug("fetched elements for merge",
		logger.String("source_branch_id", sourceBranchID.String()),
		logger.String("target_branch_id", targetBranchID.String()),
		logger.Int("source_count", len(sourceElements)),
		logger.Int("target_count", len(targetElements)),
	)

	return sourceElements, targetElements, nil
}

// DeleteBranchElements deletes all elements in a branch.
// Used when deleting a branch or resetting it.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - branchID: UUID of the branch whose elements should be deleted
//
// Returns:
//   - int: Number of elements deleted
//   - error: nil on success, error if deletion failed
func (c *SceneClient) DeleteBranchElements(ctx context.Context, branchID uuid.UUID) (int, error) {
	if c == nil || c.client == nil {
		return 0, fmt.Errorf("scene client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, sceneCallTimeout)
	defer cancel()

	// Fetch elements first
	elements, err := c.GetElements(ctx, branchID)
	if err != nil {
		return 0, fmt.Errorf("get elements: %w", err)
	}

	if len(elements) == 0 {
		return 0, nil
	}

	// Delete each element
	deletedCount := 0
	for _, el := range elements {
		_, err := c.client.DeleteElement(ctx, &scenepb.DeleteElementRequest{
			Id: el.Id,
		})

		if err != nil {
			c.log.Warn("failed to delete element",
				logger.Err(err),
				logger.String("element_id", el.Id),
				logger.String("branch_id", branchID.String()),
			)
			continue
		}

		deletedCount++
	}

	c.log.Info("branch elements deleted",
		logger.String("branch_id", branchID.String()),
		logger.Int("total_elements", len(elements)),
		logger.Int("deleted_elements", deletedCount),
	)

	return deletedCount, nil
}

// =============================================================================
// Scene Operations
// =============================================================================

// GetScene retrieves scene information.
// Used to validate scene exists when creating branches.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sceneID: UUID of the scene to retrieve
//
// Returns:
//   - *scenepb.Scene: Scene information
//   - error: nil on success, error if retrieval failed
func (c *SceneClient) GetScene(ctx context.Context, sceneID uuid.UUID) (*scenepb.Scene, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("scene client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, sceneCallTimeout)
	defer cancel()

	resp, err := c.client.GetScene(ctx, &scenepb.GetSceneRequest{
		Id: sceneID.String(),
	})
	if err != nil {
		c.log.Warn("failed to get scene from Scene Service",
			logger.Err(err),
			logger.String("scene_id", sceneID.String()),
		)
		return nil, fmt.Errorf("get scene: %w", err)
	}

	return resp.Scene, nil
}

// CheckSceneCompliance checks if a branch meets compliance rules.
// Used before allowing branch merge or approval.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sceneID: UUID of the scene
//   - branchID: UUID of the branch to check
//
// Returns:
//   - bool: true if compliant, false if violations exist
//   - []string: List of violation messages (if any)
//   - error: nil on success, error if check failed
func (c *SceneClient) CheckSceneCompliance(ctx context.Context, sceneID, branchID uuid.UUID) (bool, []string, error) {
	if c == nil || c.client == nil {
		return false, nil, fmt.Errorf("scene client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, sceneCallTimeout)
	defer cancel()

	resp, err := c.client.CheckCompliance(ctx, &scenepb.CheckComplianceRequest{
		SceneId:  sceneID.String(),
		BranchId: branchID.String(),
	})
	if err != nil {
		c.log.Warn("failed to check compliance",
			logger.Err(err),
			logger.String("scene_id", sceneID.String()),
			logger.String("branch_id", branchID.String()),
		)
		return false, nil, fmt.Errorf("check compliance: %w", err)
	}

	// Extract violation messages
	violations := make([]string, 0, len(resp.Violations))
	for _, v := range resp.Violations {
		violations = append(violations, fmt.Sprintf("[%s] %s: %s", v.Severity, v.Title, v.Description))
	}

	c.log.Debug("compliance check completed",
		logger.String("scene_id", sceneID.String()),
		logger.String("branch_id", branchID.String()),
		logger.Bool("is_compliant", resp.IsCompliant),
		logger.Int("violations", len(resp.Violations)),
	)

	return resp.IsCompliant, violations, nil
}

