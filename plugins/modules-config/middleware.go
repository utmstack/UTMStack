package main

import (
	"context"
	"crypto/subtle"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func HttpMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		internalKey := c.GetHeader("internal-key")
		if internalKey == "" {
			e := catcher.Error("missing internal-key", fmt.Errorf("missing internal-key"), map[string]any{
				"status": 404,
			})
			e.GinError(c)
			return
		}

		if internalKey != InternalKey {
			e := catcher.Error("internal key does not match", fmt.Errorf("internal key does not match"), map[string]any{
				"status": 403,
			})
			e.GinError(c)
			return
		}
		c.Next()
	}
}

func GrpcUniMiddleware(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := authFromContext(ctx); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func GrpcStreamMiddleware(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := authFromContext(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}

func authFromContext(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "metadata is not provided")
	}

	internalKey := md.Get("internal-key")
	if len(internalKey) == 0 {
		return status.Error(codes.Unauthenticated, "internal key is not provided")
	}

	if subtle.ConstantTimeCompare([]byte(internalKey[0]), []byte(InternalKey)) != 1 {
		return status.Error(codes.PermissionDenied, "internal key does not match")
	}

	return nil
}
