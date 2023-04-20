package v1

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	pb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type deleterMock struct {
	DeleteFn func(userID string, slugs []string) error
}

func (m *deleterMock) Delete(userID string, slugs []string) error {
	if m != nil && m.DeleteFn != nil {
		return m.DeleteFn(userID, slugs)
	}
	return fmt.Errorf("unable to delete URLs")
}

func TestURLsDeleterService_DelURLs(t *testing.T) {
	type services struct {
		deleter deleter
	}
	type want struct {
		code codes.Code
	}
	tests := []struct {
		serv   services
		in     *pb.DelURLsRequest
		userID string
		name   string
		want   want
	}{
		{
			name:   "Delete URLs, status code: Ok",
			userID: "1",
			in: &pb.DelURLsRequest{
				Slugs: []string{
					"slug1",
					"slug2",
					"slug3",
				},
			},
			want: want{
				code: codes.OK,
			},
			serv: services{
				deleter: &deleterMock{
					DeleteFn: func(userID string, slugs []string) error {
						return nil
					},
				},
			},
		},
		{
			name:   "No slugs for deleting, status code: InvalidArgument",
			userID: "1",
			in: &pb.DelURLsRequest{
				Slugs: []string{},
			},
			want: want{
				code: codes.InvalidArgument,
			},
		},
		{
			name:   "Delete error, status code: Internal",
			userID: "1",
			in: &pb.DelURLsRequest{
				Slugs: []string{
					"slug1",
					"slug2",
					"slug3",
				},
			},
			want: want{
				code: codes.Internal,
			},
			serv: services{
				deleter: &deleterMock{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				urlsDeleterClient(context.Background(), tt.serv.deleter)
			defer closer()

			reqCtx := metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("user_id", tt.userID))
			_, err := client.DelURLs(reqCtx, tt.in)
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

func urlsDeleterClient(
	ctx context.Context,
	del deleter,
) (pb.URLsDeleterClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterURLsDeleterServer(
		baseServer,
		newURLsDeleterService(del),
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

	client := pb.NewURLsDeleterClient(conn)
	return client, closer
}
