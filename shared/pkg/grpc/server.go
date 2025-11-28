// Package grpc provides gRPC server and client utilities for Granula microservices.
//
// Features:
// - Server configuration with interceptors
// - Logging, recovery, and tracing interceptors
// - Health checks
// - Graceful shutdown
//
// Example:
//
//	srv, err := grpc.NewServer(grpc.ServerConfig{
//	    Port: 50051,
//	    Logger: log,
//	})
//	pb.RegisterMyServiceServer(srv.Server(), &myService{})
//	srv.Start()
package grpc

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// ServerConfig holds gRPC server configuration.
type ServerConfig struct {
	// Port to listen on.
	Port int

	// Logger for request logging.
	Logger *logger.Logger

	// ServiceName for health checks and logging.
	ServiceName string

	// EnableReflection enables gRPC reflection for debugging.
	EnableReflection bool

	// MaxRecvMsgSize is the maximum message size in bytes.
	MaxRecvMsgSize int

	// MaxSendMsgSize is the maximum message size in bytes.
	MaxSendMsgSize int

	// ConnectionTimeout for client connections.
	ConnectionTimeout time.Duration

	// CustomInterceptors are additional unary interceptors.
	CustomInterceptors []grpc.UnaryServerInterceptor

	// CustomStreamInterceptors are additional stream interceptors.
	CustomStreamInterceptors []grpc.StreamServerInterceptor
}

// DefaultServerConfig returns sensible defaults.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:              50051,
		EnableReflection:  true,
		MaxRecvMsgSize:    16 * 1024 * 1024, // 16MB
		MaxSendMsgSize:    16 * 1024 * 1024, // 16MB
		ConnectionTimeout: 30 * time.Second,
	}
}

// Server wraps a gRPC server with additional functionality.
type Server struct {
	server       *grpc.Server
	healthServer *health.Server
	listener     net.Listener
	config       ServerConfig
	log          *logger.Logger
}

// NewServer creates a new gRPC server.
func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Logger == nil {
		cfg.Logger = logger.Global()
	}

	// Build interceptor chains
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		recoveryInterceptor(cfg.Logger),
		loggingInterceptor(cfg.Logger),
		contextInterceptor(),
	}
	unaryInterceptors = append(unaryInterceptors, cfg.CustomInterceptors...)

	streamInterceptors := []grpc.StreamServerInterceptor{
		streamRecoveryInterceptor(cfg.Logger),
		streamLoggingInterceptor(cfg.Logger),
	}
	streamInterceptors = append(streamInterceptors, cfg.CustomStreamInterceptors...)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}

	if cfg.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize))
	}
	if cfg.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.MaxSendMsgSize(cfg.MaxSendMsgSize))
	}
	if cfg.ConnectionTimeout > 0 {
		opts = append(opts, grpc.ConnectionTimeout(cfg.ConnectionTimeout))
	}

	server := grpc.NewServer(opts...)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Enable reflection for debugging
	if cfg.EnableReflection {
		reflection.Register(server)
	}

	return &Server{
		server:       server,
		healthServer: healthServer,
		config:       cfg,
		log:          cfg.Logger,
	}, nil
}

// Server returns the underlying gRPC server for service registration.
func (s *Server) Server() *grpc.Server {
	return s.server
}

// SetServingStatus sets the health status for a service.
func (s *Server) SetServingStatus(service string, serving bool) {
	status := grpc_health_v1.HealthCheckResponse_SERVING
	if !serving {
		status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}
	s.healthServer.SetServingStatus(service, status)
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.log.Info("gRPC server starting",
		logger.String("addr", addr),
		logger.String("service", s.config.ServiceName),
	)

	// Set service as serving
	if s.config.ServiceName != "" {
		s.SetServingStatus(s.config.ServiceName, true)
	}
	s.SetServingStatus("", true) // Overall health

	return s.server.Serve(listener)
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	s.log.Info("gRPC server stopping")

	// Set service as not serving
	if s.config.ServiceName != "" {
		s.SetServingStatus(s.config.ServiceName, false)
	}
	s.SetServingStatus("", false)

	s.server.GracefulStop()
}

// ForceStop immediately stops the server.
func (s *Server) ForceStop() {
	s.server.Stop()
}

// -----------------------------------------------------------------------------
// Interceptors
// -----------------------------------------------------------------------------

// recoveryInterceptor recovers from panics.
func recoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					logger.F("panic", r),
					logger.String("method", info.FullMethod),
					logger.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// loggingInterceptor logs requests and responses.
func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract metadata
		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		var requestID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-request-id"); len(values) > 0 {
				requestID = values[0]
			}
		}

		// Call handler
		resp, err := handler(ctx, req)

		// Log request
		duration := time.Since(start)
		fields := []zap.Field{
			logger.String("method", info.FullMethod),
			logger.Duration("duration", duration.Milliseconds()),
			logger.String("client_ip", clientIP),
		}

		if requestID != "" {
			fields = append(fields, logger.String("request_id", requestID))
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields, logger.String("code", st.Code().String()))
			fields = append(fields, logger.String("error", st.Message()))
			log.Warn("gRPC request failed", fields...)
		} else {
			fields = append(fields, logger.String("code", "OK"))
			log.Info("gRPC request", fields...)
		}

		return resp, err
	}
}

// contextInterceptor adds request context values.
func contextInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract values from metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-request-id"); len(values) > 0 {
				ctx = context.WithValue(ctx, "request_id", values[0])
			}
			if values := md.Get("x-user-id"); len(values) > 0 {
				ctx = context.WithValue(ctx, "user_id", values[0])
			}
			if values := md.Get("x-trace-id"); len(values) > 0 {
				ctx = context.WithValue(ctx, "trace_id", values[0])
			}
		}

		return handler(ctx, req)
	}
}

// streamRecoveryInterceptor recovers from panics in streams.
func streamRecoveryInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("stream panic recovered",
					logger.F("panic", r),
					logger.String("method", info.FullMethod),
					logger.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(srv, ss)
	}
}

// streamLoggingInterceptor logs stream requests.
func streamLoggingInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		log.Info("gRPC stream started",
			logger.String("method", info.FullMethod),
			logger.Bool("client_stream", info.IsClientStream),
			logger.Bool("server_stream", info.IsServerStream),
		)

		err := handler(srv, ss)

		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.Warn("gRPC stream failed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration.Milliseconds()),
				logger.String("code", st.Code().String()),
				logger.String("error", st.Message()),
			)
		} else {
			log.Info("gRPC stream completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration.Milliseconds()),
			)
		}

		return err
	}
}
