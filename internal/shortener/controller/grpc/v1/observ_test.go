package v1

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	pb "github.com/alukart32/shortener-url/pkg/proto/v1/observ"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type pingerMock struct {
	PingFn func() error
}

func (m *pingerMock) Ping() error {
	if m != nil && m.PingFn != nil {
		return m.PingFn()
	}
	return fmt.Errorf("unable to ping")
}

func TestObservService_PingPostgres(t *testing.T) {
	type services struct {
		ping pinger
	}
	tests := []struct {
		serv     services
		name     string
		wantCode codes.Code
	}{
		{
			name:     "Ping, status code: Ok",
			wantCode: codes.OK,
			serv: services{
				ping: &pingerMock{
					PingFn: func() error { return nil },
				},
			},
		},
		{
			name:     "Ping error, status code: Internal",
			wantCode: codes.Internal,
			serv: services{
				ping: &pingerMock{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				observClient(context.Background(), tt.serv.ping)
			defer closer()

			_, err := client.PingPostgres(context.Background(), &pb.PingPostgresRequest{})
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.EqualValues(t, tt.wantCode, e.Code(),
						"Expected status code: %d, got %d", tt.wantCode, e.Code())
					return
				} else {
					t.Fatalf("failed to parse: %v", err)
				}
			}
		})
	}
}

func observClient(
	ctx context.Context,
	ping pinger,
) (pb.ObservabilityClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterObservabilityServer(
		baseServer,
		newObservService(ping),
	)
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := pb.NewObservabilityClient(conn)
	return client, closer
}
