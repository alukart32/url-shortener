package v1

import (
	"context"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	pb "github.com/alukart32/shortener-url/pkg/proto/v1/stat"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// statProvider defines the shortened URLs statistics provider.
type statProvider interface {
	Stat(context.Context) (models.Stat, error)
}

// statService is a representation of the proto StatisticsServer.
type statService struct {
	pb.UnimplementedStatisticsServer
	provider statProvider
}

// newStatService returns a new statService.
func newStatService(provider statProvider) *statService {
	return &statService{provider: provider}
}

// Stat collects shortened URLs statistics.
func (s *statService) Stat(ctx context.Context, _ *pb.StatRequest) (*pb.StatResponse, error) {
	stat, err := s.provider.Stat(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var response pb.StatResponse
	response.UrlsCount = int32(stat.URLs)
	response.UsersCount = int32(stat.Users)

	return &response, nil
}
