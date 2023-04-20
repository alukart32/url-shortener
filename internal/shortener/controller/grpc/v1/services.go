// Package v1 provides services for gRPC API v1 routes.
package v1

import (
	"context"

	"github.com/alukart32/shortener-url/internal/shortener/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	observpb "github.com/alukart32/shortener-url/pkg/proto/v1/observ"
	statpb "github.com/alukart32/shortener-url/pkg/proto/v1/stat"
	urlspb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
)

// RegServices adds gRPC services to the gRPC server.
func RegServices(srv *grpc.Server, servs *services.Services) {
	// Set urls services.
	urlspb.RegisterURLsShortenerServer(srv, newURLsShortenerService(
		servs.Provider,
		servs.Shortener,
	))
	urlspb.RegisterURLsProviderServer(srv, newURLsProviderService(
		servs.Provider,
	))
	urlspb.RegisterURLsDeleterServer(srv, newURLsDeleterService(servs.Deleter))

	// Set stat service.
	statpb.RegisterStatisticsServer(srv, newStatService(servs.Statistic))

	// Set observ services.
	observpb.RegisterObservabilityServer(srv, newObservService(servs.PostgresPinger))
}

// MethodsForAuthSkip returns a list of gRPC methods for auth skip.
func MethodsForAuthSkip() []string {
	var skipMethods []string
	skipMethods = append(
		skipMethods,
		urlspb.URLsProvider_GetShortenedURL_FullMethodName,
		observpb.Observability_PingPostgres_FullMethodName,
		statpb.Statistics_Stat_FullMethodName,
	)

	return skipMethods
}

// getUserIDFromCtx gets userID from the context of the method request.
func getUserIDFromCtx(ctx context.Context) string {
	var userID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("user_id")
		if len(values) > 0 {
			userID = values[0]
		}
	}
	return userID
}
