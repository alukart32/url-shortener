// Package grpcauth provides auth options for gRPC server.
package grpcauth

import (
	"context"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthOpts defines grpc server auth options.
type AuthOpts struct {
	AuthFn      auth.AuthFunc
	SkipMethods selector.Matcher
}

// NewAuthOpts returns a new AuthOpts.
func NewAuthOpts(jwtAuthProvider authProvider, passMethods []string) *AuthOpts {
	return &AuthOpts{
		AuthFn:      authFunc(jwtAuthProvider),
		SkipMethods: selector.MatchFunc(skipSelector(passMethods)),
	}
}

type authProvider interface {
	NewToken(userID string) (string, error)
	VerifyToken(tokenString string) (string, error)
}

// authFunc authenticates incoming methods call.
func authFunc(jwtAuthProvider authProvider) auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		var ctxWithUserID = func(ctx context.Context, userID string) context.Context {
			md := metadata.New(map[string]string{"user_id": userID})
			return metadata.NewIncomingContext(ctx, md)
		}
		var setAuthHeader = func(token string) {
			md := metadata.Pairs("authorization", "bearer "+token)
			grpc.SetHeader(ctx, md)
		}

		token, err := auth.AuthFromMD(ctx, "bearer")
		if err == nil {
			userID, err := jwtAuthProvider.VerifyToken(token)
			if err != nil {
				return ctx, status.Errorf(codes.Unauthenticated, "failed to parse token")
			}
			return ctxWithUserID(ctx, userID), nil
		}

		// Generate a new userUUID.
		userUUID, err := uuid.NewUUID()
		if err != nil {
			return ctx, status.Errorf(codes.Internal, "failed to gen userID: %v", err)
		}
		token, err = jwtAuthProvider.NewToken(userUUID.String())
		if err != nil {
			return ctx, status.Errorf(codes.Internal, "failed to gen a new JWT token: %v", err)
		}
		setAuthHeader(token)
		return ctxWithUserID(ctx, userUUID.String()), nil
	}
}

// skipSelector skips method call to process by grpc server interceptor.
func skipSelector(passMethods []string) func(ctx context.Context, c interceptors.CallMeta) bool {
	return func(_ context.Context, c interceptors.CallMeta) bool {
		for _, v := range passMethods {
			if c.FullMethod() == v {
				return false
			}
		}

		return true
	}
}
