// =============================================================================
// Package grpc provides gRPC client connections to backend microservices.
// =============================================================================
// This package manages all gRPC connections to backend services, providing
// a centralized way to create, configure, and manage client connections.
//
// Usage:
//
//	clients, err := grpc.NewClients(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer clients.Close()
//
// =============================================================================
package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/xiiisorate/granula_api/api-gateway/internal/config"
)

// Clients holds all gRPC client connections.
type Clients struct {
	// AuthConn is the connection to Auth Service.
	AuthConn *grpc.ClientConn

	// UserConn is the connection to User Service.
	UserConn *grpc.ClientConn

	// WorkspaceConn is the connection to Workspace Service.
	WorkspaceConn *grpc.ClientConn

	// FloorPlanConn is the connection to Floor Plan Service.
	FloorPlanConn *grpc.ClientConn

	// SceneConn is the connection to Scene Service.
	SceneConn *grpc.ClientConn

	// BranchConn is the connection to Branch Service.
	BranchConn *grpc.ClientConn

	// AIConn is the connection to AI Service.
	AIConn *grpc.ClientConn

	// ComplianceConn is the connection to Compliance Service.
	ComplianceConn *grpc.ClientConn

	// RequestConn is the connection to Request Service.
	RequestConn *grpc.ClientConn

	// NotificationConn is the connection to Notification Service.
	NotificationConn *grpc.ClientConn
}

// NewClients creates and establishes connections to all backend services.
func NewClients(cfg *config.Config) (*Clients, error) {
	clients := &Clients{}
	var err error

	// Connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Connect to Auth Service
	clients.AuthConn, err = grpc.NewClient(cfg.AuthServiceAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	// Connect to User Service
	clients.UserConn, err = grpc.NewClient(cfg.UserServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	// Connect to Workspace Service
	clients.WorkspaceConn, err = grpc.NewClient(cfg.WorkspaceServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to workspace service: %w", err)
	}

	// Connect to Floor Plan Service
	clients.FloorPlanConn, err = grpc.NewClient(cfg.FloorPlanServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to floorplan service: %w", err)
	}

	// Connect to Scene Service
	clients.SceneConn, err = grpc.NewClient(cfg.SceneServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to scene service: %w", err)
	}

	// Connect to Branch Service
	clients.BranchConn, err = grpc.NewClient(cfg.BranchServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to branch service: %w", err)
	}

	// Connect to AI Service
	clients.AIConn, err = grpc.NewClient(cfg.AIServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to ai service: %w", err)
	}

	// Connect to Compliance Service
	clients.ComplianceConn, err = grpc.NewClient(cfg.ComplianceServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to compliance service: %w", err)
	}

	// Connect to Request Service
	clients.RequestConn, err = grpc.NewClient(cfg.RequestServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to request service: %w", err)
	}

	// Connect to Notification Service
	clients.NotificationConn, err = grpc.NewClient(cfg.NotificationServiceAddr, opts...)
	if err != nil {
		clients.Close()
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}

	return clients, nil
}

// Close closes all gRPC connections.
func (c *Clients) Close() {
	if c.AuthConn != nil {
		c.AuthConn.Close()
	}
	if c.UserConn != nil {
		c.UserConn.Close()
	}
	if c.WorkspaceConn != nil {
		c.WorkspaceConn.Close()
	}
	if c.FloorPlanConn != nil {
		c.FloorPlanConn.Close()
	}
	if c.SceneConn != nil {
		c.SceneConn.Close()
	}
	if c.BranchConn != nil {
		c.BranchConn.Close()
	}
	if c.AIConn != nil {
		c.AIConn.Close()
	}
	if c.ComplianceConn != nil {
		c.ComplianceConn.Close()
	}
	if c.RequestConn != nil {
		c.RequestConn.Close()
	}
	if c.NotificationConn != nil {
		c.NotificationConn.Close()
	}
}

// HealthCheck checks if all services are reachable.
func (c *Clients) HealthCheck(ctx context.Context) map[string]bool {
	results := make(map[string]bool)

	results["auth"] = c.AuthConn.GetState().String() == "READY"
	results["user"] = c.UserConn.GetState().String() == "READY"
	results["workspace"] = c.WorkspaceConn.GetState().String() == "READY"
	results["floorplan"] = c.FloorPlanConn.GetState().String() == "READY"
	results["scene"] = c.SceneConn.GetState().String() == "READY"
	results["branch"] = c.BranchConn.GetState().String() == "READY"
	results["ai"] = c.AIConn.GetState().String() == "READY"
	results["compliance"] = c.ComplianceConn.GetState().String() == "READY"
	results["request"] = c.RequestConn.GetState().String() == "READY"
	results["notification"] = c.NotificationConn.GetState().String() == "READY"

	return results
}

