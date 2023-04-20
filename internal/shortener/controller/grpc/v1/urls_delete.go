package v1

import (
	"context"

	pb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// deleter defines the shortened URLs deleter by slugs.
type deleter interface {
	Delete(userID string, slugs []string) error
}

// urlsDeleterService is a representation of the proto URLsDeleterServer.
type urlsDeleterService struct {
	pb.UnimplementedURLsDeleterServer
	deleter deleter
}

// newURLsDeleterService returns a new urlsDeleterService.
func newURLsDeleterService(deleter deleter) *urlsDeleterService {
	return &urlsDeleterService{deleter: deleter}
}

// DelURLs marks shortened URLs as deleted.
func (s *urlsDeleterService) DelURLs(ctx context.Context, in *pb.DelURLsRequest) (*pb.DelURLsResponse, error) {
	if len(in.Slugs) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "empty slugs")
	}

	userID := getUserIDFromCtx(ctx)
	if err := s.deleter.Delete(userID, in.Slugs); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DelURLsResponse{}, nil
}
