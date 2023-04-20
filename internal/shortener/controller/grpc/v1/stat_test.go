package v1

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	pb "github.com/alukart32/shortener-url/pkg/proto/v1/stat"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type statProviderMock struct {
	StatFn func(context.Context) (models.Stat, error)
}

func (m *statProviderMock) Stat(ctx context.Context) (models.Stat, error) {
	if m != nil && m.StatFn != nil {
		return m.StatFn(ctx)
	}
	return models.Stat{}, fmt.Errorf("unable to short URL")
}

func TestStatService_Stat(t *testing.T) {
	type services struct {
		stat statProvider
	}
	type want struct {
		data models.Stat
		code codes.Code
	}
	tests := []struct {
		serv services
		name string
		want want
	}{
		{
			name: "Delete URLs, status code: Ok",
			want: want{
				data: models.Stat{
					URLs:  7,
					Users: 2,
				},
				code: codes.OK,
			},
			serv: services{
				stat: &statProviderMock{
					StatFn: func(ctx context.Context) (models.Stat, error) {
						return models.Stat{URLs: 7, Users: 2}, nil
					},
				},
			},
		},
		{
			name: "Stat error, status code: Internal",
			want: want{
				code: codes.Internal,
			},
			serv: services{
				stat: &statProviderMock{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				statClient(context.Background(), tt.serv.stat)
			defer closer()

			_, err := client.Stat(context.Background(), &pb.StatRequest{})
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.EqualValues(t, tt.want.code, e.Code(),
						"Expected status code: %d, got %d", tt.want.code, e.Code())
					return
				} else {
					t.Fatalf("failed to parse: %v", err)
				}
			}
		})
	}
}

func statClient(
	ctx context.Context,
	provider statProvider,
) (pb.StatisticsClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterStatisticsServer(
		baseServer,
		newStatService(provider),
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

	client := pb.NewStatisticsClient(conn)
	return client, closer
}
