package v1

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/services/shorturl"
	pb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type shortenerMock struct {
	ShortFn func(context.Context, models.URL) (string, error)
	BatchFn func(context.Context, []models.URL) ([]models.ShortenedURL, error)
}

func (m *shortenerMock) Short(ctx context.Context, url models.URL) (string, error) {
	if m != nil && m.ShortFn != nil {
		return m.ShortFn(ctx, url)
	}
	return "", fmt.Errorf("unable to short URL")
}

func (m *shortenerMock) Batch(ctx context.Context, urls []models.URL) ([]models.ShortenedURL, error) {
	if m != nil && m.BatchFn != nil {
		return m.BatchFn(ctx, urls)
	}
	return nil, fmt.Errorf("unable to batch URLs")
}

type getterByURLMock struct {
	GetByURLFn func(context.Context, string) (models.ShortenedURL, error)
}

func (m *getterByURLMock) GetByURL(ctx context.Context, raw string) (models.ShortenedURL, error) {
	if m != nil && m.GetByURLFn != nil {
		return m.GetByURLFn(ctx, raw)
	}
	return models.ShortenedURL{}, fmt.Errorf("unable to short URL")
}

func TestURLsShortenerService_ShortURL(t *testing.T) {
	type services struct {
		shortsrv shortener
		getter   getterByURL
	}
	type req struct {
		in     *pb.ShortURLRequest
		userID string
	}
	type want struct {
		data *pb.ShortURLResponse
		code codes.Code
	}
	tests := []struct {
		req  req
		want want
		serv services
		name string
	}{
		{
			name: "New URL, status code: Ok",
			req: req{
				userID: "1",
				in: &pb.ShortURLRequest{
					Raw: "http://demo.com",
				},
			},
			want: want{
				code: codes.OK,
				data: &pb.ShortURLResponse{
					ShortUrl: "https://localhost:8080/slug1",
				},
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(ctx context.Context, u models.URL) (string, error) {
						return "https://localhost:8080/slug1", nil
					},
				},
			},
		},
		{
			name: "Existed URL, status code: AlreadyExists",
			req: req{
				userID: "1",
				in: &pb.ShortURLRequest{
					Raw: "http://demo.com",
				},
			},
			want: want{
				code: codes.AlreadyExists,
				data: &pb.ShortURLResponse{
					ShortUrl: "https://localhost:8080/tmp_slug",
				},
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "", shorturl.ErrUniqueViolation
					},
				},
				getter: &getterByURLMock{
					GetByURLFn: func(_ context.Context, s string) (models.ShortenedURL, error) {
						return models.NewShortenedURL("1", "1", "https//go-dev",
							"tmp_slug", "http//localhost:8080/tmp_slug"), nil
					},
				},
			},
		},
		{
			name: "Empty request, status code: InvalidArgument",
			req: req{
				userID: "1",
				in: &pb.ShortURLRequest{
					Raw: "",
				},
			},
			want: want{
				code: codes.InvalidArgument,
			},
			serv: services{
				shortsrv: &shortenerMock{
					ShortFn: func(_ context.Context, u models.URL) (string, error) {
						return "", shorturl.ErrInvalidCreation
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				urlsShortenerClient(context.Background(), tt.serv.shortsrv, tt.serv.getter)
			defer closer()

			reqCtx := metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("user_id", tt.req.userID))
			resp, err := client.ShortURL(reqCtx, tt.req.in)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.EqualValues(t, tt.want.code, e.Code(),
						"Expected status code: %d, got %d", tt.want.code, e.Code())
					return
				} else {
					t.Fatalf("failed to parse: %v", err)
				}
			}

			assert.EqualValues(t, tt.want.data.ShortUrl, resp.ShortUrl,
				"Expected shortURL: %s, got %s", tt.want.data.ShortUrl, resp.ShortUrl)
		})
	}
}

func TestURLsShortenerService_BatchURLs(t *testing.T) {
	type services struct {
		shortsrv shortener
	}
	type want struct {
		data *pb.BatchURLsResponse
		code codes.Code
	}
	tests := []struct {
		want   want
		serv   services
		req    *pb.BatchURLsRequest
		userID string
		name   string
	}{
		{
			name:   "List URLs for userID 1, status code: Ok",
			userID: "1",
			req: &pb.BatchURLsRequest{
				Url: []*pb.BatchURLsRequest_URL{
					{
						CorrId: "1",
						Raw:    "http://demo.com/1",
					},
					{
						CorrId: "2",
						Raw:    "http://demo.com/2",
					},
					{
						CorrId: "3",
						Raw:    "http://demo.com/3",
					},
					{
						CorrId: "4",
						Raw:    "http://demo.com/4",
					},
					{
						CorrId: "5",
						Raw:    "http://demo.com/5",
					},
					{
						CorrId: "6",
						Raw:    "http://demo.com/6",
					},
					{
						CorrId: "7",
						Raw:    "http://demo.com/7",
					},
				},
			},
			want: want{
				data: &pb.BatchURLsResponse{
					BatchedUrls: []*pb.BatchURLsResponse_URL{
						{
							CorrId:   "1",
							ShortUrl: "http://localhost:8080/slug1",
						},
						{
							CorrId:   "2",
							ShortUrl: "http://localhost:8080/slug2",
						},
						{
							CorrId:   "3",
							ShortUrl: "http://localhost:8080/slug3",
						},
						{
							CorrId:   "4",
							ShortUrl: "http://localhost:8080/slug4",
						},
						{
							CorrId:   "5",
							ShortUrl: "http://localhost:8080/slug5",
						},
						{
							CorrId:   "6",
							ShortUrl: "http://localhost:8080/slug6",
						},
						{
							CorrId:   "7",
							ShortUrl: "http://localhost:8080/slug7",
						},
					},
				},
				code: codes.OK,
			},
			serv: services{
				shortsrv: &shortenerMock{
					BatchFn: func(ctx context.Context, urls []models.URL) ([]models.ShortenedURL, error) {
						records := []models.ShortenedURL{
							{
								CorrID: "1",
								Value:  "http://localhost:8080/slug1",
							},
							{
								CorrID: "2",
								Value:  "http://localhost:8080/slug2",
							},
							{
								CorrID: "3",
								Value:  "http://localhost:8080/slug3",
							},
							{
								CorrID: "4",
								Value:  "http://localhost:8080/slug4",
							},
							{
								CorrID: "5",
								Value:  "http://localhost:8080/slug5",
							},
							{
								CorrID: "6",
								Value:  "http://localhost:8080/slug6",
							},
							{
								CorrID: "7",
								Value:  "http://localhost:8080/slug7",
							},
						}
						return records, nil
					},
				},
			},
		},
		{
			name:   "No URLs for batching, status code: InvalidArgument",
			userID: "1",
			req: &pb.BatchURLsRequest{
				Url: []*pb.BatchURLsRequest_URL{},
			},
			want: want{
				code: codes.InvalidArgument,
			},
			serv: services{
				shortsrv: &shortenerMock{
					BatchFn: func(ctx context.Context, u []models.URL) ([]models.ShortenedURL, error) {
						return nil, shorturl.ErrEmptyBatch
					},
				},
			},
		},
		{
			name:   "Invalid URLs for batching, status code: InvalidArgument",
			userID: "1",
			req: &pb.BatchURLsRequest{
				Url: []*pb.BatchURLsRequest_URL{
					{
						CorrId: "1",
						Raw:    "http//demo.com/1",
					},
					{
						CorrId: "2",
						Raw:    "httpa://demo.com/2",
					},
					{
						CorrId: "3",
						Raw:    "http:demo.com/3",
					},
				},
			},
			want: want{
				code: codes.InvalidArgument,
			},
			serv: services{
				shortsrv: &shortenerMock{
					BatchFn: func(ctx context.Context, u []models.URL) ([]models.ShortenedURL, error) {
						return nil, shorturl.ErrInvalidCreation
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				urlsShortenerClient(context.Background(), tt.serv.shortsrv, nil)
			defer closer()

			reqCtx := metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("user_id", tt.userID))
			resp, err := client.BatchURLs(reqCtx, tt.req)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.EqualValues(t, tt.want.code, e.Code(),
						"Expected status code: %d, got %d", tt.want.code, e.Code())
					return
				} else {
					t.Fatalf("failed to parse: %v", err)
				}
			}

			assert.EqualValues(t, len(tt.want.data.BatchedUrls), len(resp.BatchedUrls))
		})
	}
}

func urlsShortenerClient(
	ctx context.Context,
	shortsrv shortener,
	getter getterByURL,
) (pb.URLsShortenerClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterURLsShortenerServer(
		baseServer,
		newURLsShortenerService(getter, shortsrv),
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

	client := pb.NewURLsShortenerClient(conn)
	return client, closer
}
