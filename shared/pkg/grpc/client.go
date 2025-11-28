// Package grpc provides gRPC client utilities for Granula microservices.

package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// ClientConfig holds gRPC client configuration.
type ClientConfig struct {
	// Target is the server address (host:port).
	Target string

	// Logger for request logging.
	Logger *logger.Logger

	// Timeout for connection establishment.
	DialTimeout time.Duration

	// RequestTimeout is the default timeout for requests.
	RequestTimeout time.Duration

	// MaxRetries is the number of retry attempts.
	MaxRetries int

	// RetryBackoff is the initial retry backoff duration.
	RetryBackoff time.Duration

	// MaxRecvMsgSize is the maximum message size in bytes.
	MaxRecvMsgSize int

	// MaxSendMsgSize is the maximum message size in bytes.
	MaxSendMsgSize int

	// Insecure disables TLS (for development).
	Insecure bool

	// KeepAliveTime is the keep-alive ping interval.
	KeepAliveTime time.Duration

	// KeepAliveTimeout is the keep-alive ping timeout.
	KeepAliveTimeout time.Duration
}

// DefaultClientConfig returns sensible defaults.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		DialTimeout:      10 * time.Second,
		RequestTimeout:   30 * time.Second,
		MaxRetries:       3,
		RetryBackoff:     100 * time.Millisecond,
		MaxRecvMsgSize:   16 * 1024 * 1024, // 16MB
		MaxSendMsgSize:   16 * 1024 * 1024, // 16MB
		Insecure:         true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}
}

// Client wraps a gRPC client connection.
type Client struct {
	conn   *grpc.ClientConn
	config ClientConfig
	log    *logger.Logger
}

// NewClient creates a new gRPC client.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.Target == "" {
		return nil, fmt.Errorf("target address is required")
	}

	if cfg.Logger == nil {
		cfg.Logger = logger.Global()
	}

	// Apply defaults
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = DefaultClientConfig().DialTimeout
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = DefaultClientConfig().RequestTimeout
	}

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.MaxSendMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepAliveTime,
			Timeout:             cfg.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithChainUnaryInterceptor(
			clientLoggingInterceptor(cfg.Logger),
			clientTimeoutInterceptor(cfg.RequestTimeout),
		),
		grpc.WithChainStreamInterceptor(
			clientStreamLoggingInterceptor(cfg.Logger),
		),
	}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", cfg.Target, err)
	}

	cfg.Logger.Info("gRPC client connected",
		logger.String("target", cfg.Target),
	)

	return &Client{
		conn:   conn,
		config: cfg,
		log:    cfg.Logger,
	}, nil
}

// Conn returns the underlying gRPC connection.
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.log.Info("gRPC client closing",
		logger.String("target", c.config.Target),
	)
	return c.conn.Close()
}

// -----------------------------------------------------------------------------
// Client Interceptors
// -----------------------------------------------------------------------------

// clientLoggingInterceptor logs client requests.
func clientLoggingInterceptor(log *logger.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)

		if err != nil {
			log.Debug("gRPC client call failed",
				logger.String("method", method),
				logger.Duration("duration", duration.Milliseconds()),
				logger.String("target", cc.Target()),
				logger.Err(err),
			)
		} else {
			log.Debug("gRPC client call",
				logger.String("method", method),
				logger.Duration("duration", duration.Milliseconds()),
				logger.String("target", cc.Target()),
			)
		}

		return err
	}
}

// clientTimeoutInterceptor adds timeout to client requests.
func clientTimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Only add timeout if not already set
		if _, ok := ctx.Deadline(); !ok && timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// clientStreamLoggingInterceptor logs client streams.
func clientStreamLoggingInterceptor(log *logger.Logger) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		log.Debug("gRPC client stream started",
			logger.String("method", method),
			logger.String("target", cc.Target()),
		)

		return streamer(ctx, desc, cc, method, opts...)
	}
}

// -----------------------------------------------------------------------------
// Context helpers
// -----------------------------------------------------------------------------

// WithRequestID adds request ID to context for outgoing calls.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
}

// WithUserID adds user ID to context for outgoing calls.
func WithUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)
}

// WithTraceID adds trace ID to context for outgoing calls.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceID)
}

// WithAuthToken adds authorization token to context.
func WithAuthToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

// PropagateMetadata propagates metadata from incoming to outgoing context.
func PropagateMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	// Propagate specific headers
	headers := []string{"x-request-id", "x-trace-id", "x-user-id", "authorization"}
	outMD := metadata.MD{}

	for _, h := range headers {
		if values := md.Get(h); len(values) > 0 {
			outMD.Set(h, values...)
		}
	}

	return metadata.NewOutgoingContext(ctx, outMD)
}
