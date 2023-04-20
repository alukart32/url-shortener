package v1

import (
	"context"

	pb "github.com/alukart32/shortener-url/pkg/proto/v1/observ"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// pinger defines the network pinger.
type pinger interface {
	Ping() error
}

// observService is a representation of the proto ObservabilityServer.
type observService struct {
	pb.UnimplementedObservabilityServer
	pinger pinger
}

// newStatService returns a new observService.
func newObservService(pinger pinger) *observService {
	return &observService{pinger: pinger}
}

// PingPostgres checks postgres health by network ping request.
func (s *observService) PingPostgres(_ context.Context, _ *pb.PingPostgresRequest) (*pb.PingPostgresResponse, error) {
	if err := s.pinger.Ping(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PingPostgresResponse{}, nil
}
