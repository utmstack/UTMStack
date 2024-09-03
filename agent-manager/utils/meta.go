package utils

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetItemsFromContext(ctx context.Context) (string, string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", "", status.Error(codes.Internal, "metadata is not provided")
	}
	id, ok := md["id"]
	if !ok || len(id) == 0 {
		return "", "", "", status.Error(codes.Unauthenticated, "id is not provided")
	}
	key, ok := md["key"]
	if !ok || len(key) == 0 {
		return "", "", "", status.Error(codes.Unauthenticated, "key is not provided")
	}
	typ, ok := md["type"]
	if !ok || len(typ) == 0 {
		return "", "", "", status.Error(codes.Unauthenticated, "type is not provided")
	}
	return id[0], key[0], strings.ToLower(typ[0]), nil
}
