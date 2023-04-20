package v1

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	pb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type providerMock struct {
	GetBySlugFn     func(context.Context, string) (models.ShortenedURL, error)
	CollectByUserFn func(context.Context, string) ([]models.ShortenedURL, error)
}

func (m *providerMock) GetBySlug(ctx context.Context, slug string) (models.ShortenedURL, error) {
	if m != nil && m.GetBySlugFn != nil {
		return m.GetBySlugFn(ctx, slug)
	}
	return models.ShortenedURL{}, fmt.Errorf("unable to get an URL using slug")
}

func (m *providerMock) CollectByUser(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	if m != nil && m.CollectByUserFn != nil {
		return m.CollectByUserFn(ctx, userID)
	}
	return nil, fmt.Errorf("unable to collect URLs.")
}

func TestURLsProviderService_GetShortenedURL(t *testing.T) {
	type services struct {
		provider provider
	}
	type want struct {
		data *pb.GetShortenedURLResponse
		code codes.Code
	}
	tests := []struct {
		want want
		serv services
		req  *pb.GetShortenedURLRequest
		name string
	}{
		{
			name: "URL by 112Sd exists, status code: Ok",
			req: &pb.GetShortenedURLRequest{
				Slug: "112Sd",
			},
			want: want{
				data: &pb.GetShortenedURLResponse{
					ShortUrl: &pb.GetShortenedURLResponse_ShortenedURL{
						UserId: "1",
						CorrId: "1",
						Raw:    "http://example.com/query_1",
						Slug:   "tmp_slug",
						Value:  "http://localhost:8080/tmp_slug",
					},
				},
				code: codes.OK,
			},
			serv: services{
				provider: &providerMock{
					GetBySlugFn: func(ctx context.Context, s string) (models.ShortenedURL, error) {
						return models.NewShortenedURL("1", "1", "http://example.com/query_1",
							"tmp_slug", "http://localhost:8080/tmp_slug"), nil
					},
				},
			},
		},
		{
			name: "URL doesn't exist, status code: NotFound",
			req: &pb.GetShortenedURLRequest{
				Slug: "112Sd",
			},
			want: want{
				code: codes.NotFound,
			},
			serv: services{
				provider: &providerMock{
					GetBySlugFn: func(ctx context.Context, s string) (models.ShortenedURL, error) {
						return models.ShortenedURL{}, nil
					},
				},
			},
		},
		{
			name: "URL by 112Sd was deleted, status code: Unknown",
			req: &pb.GetShortenedURLRequest{
				Slug: "112Sd",
			},
			want: want{
				data: &pb.GetShortenedURLResponse{
					ShortUrl: &pb.GetShortenedURLResponse_ShortenedURL{
						UserId:    "1",
						CorrId:    "1",
						Raw:       "http://example.com/query_1",
						Slug:      "tmp_slug",
						Value:     "http://localhost:8080/tmp_slug",
						IsDeleted: true,
					},
				},
				code: codes.Unknown,
			},
			serv: services{
				provider: &providerMock{
					GetBySlugFn: func(ctx context.Context, s string) (models.ShortenedURL, error) {
						url := models.NewShortenedURL("1", "1", "http://example.com/query_1",
							"tmp_slug", "http://localhost:8080/tmp_slug")
						url.IsDeleted = true
						return url, nil
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				urlsProviderClient(context.Background(), tt.serv.provider)
			defer closer()

			resp, err := client.GetShortenedURL(context.Background(), tt.req)
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

func TestURLsProviderService_ListURLs(t *testing.T) {
	type services struct {
		provider provider
	}
	type want struct {
		data *pb.ListURLsResponse
		code codes.Code
	}
	tests := []struct {
		want   want
		serv   services
		userID string
		name   string
	}{
		{
			name:   "List URLs for userID 1, status code: Ok",
			userID: "1",
			want: want{
				data: &pb.ListURLsResponse{
					CollectedUrls: []*pb.ListURLsResponse_URL{
						{
							Raw:      "http://demo.com/1",
							ShortUrl: "http://localhost:8080/slug1",
						},
						{
							Raw:      "http://demo.com/2",
							ShortUrl: "http://localhost:8080/slug2",
						},
						{
							Raw:      "http://demo.com/3",
							ShortUrl: "http://localhost:8080/slug3",
						},
						{
							Raw:      "http://demo.com/4",
							ShortUrl: "http://localhost:8080/slug4",
						},
						{
							Raw:      "http://demo.com/5",
							ShortUrl: "http://localhost:8080/slug5",
						},
						{
							Raw:      "http://demo.com/6",
							ShortUrl: "http://localhost:8080/slug6",
						},
						{
							Raw:      "http://demo.com/7",
							ShortUrl: "http://localhost:8080/slug7",
						},
					},
				},
				code: codes.OK,
			},
			serv: services{
				provider: &providerMock{
					CollectByUserFn: func(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
						records := []models.ShortenedURL{
							{
								Raw:   "http://demo.com/1",
								Value: "http://localhost:8080/slug1",
							},
							{
								Raw:   "http://demo.com/2",
								Value: "http://localhost:8080/slug2",
							},
							{
								Raw:   "http://demo.com/3",
								Value: "http://localhost:8080/slug3",
							},
							{
								Raw:   "http://demo.com/4",
								Value: "http://localhost:8080/slug4",
							},
							{
								Raw:   "http://demo.com/5",
								Value: "http://localhost:8080/slug5",
							},
							{
								Raw:   "http://demo.com/6",
								Value: "http://localhost:8080/slug6",
							},
							{
								Raw:   "http://demo.com/7",
								Value: "http://localhost:8080/slug7",
							},
						}
						return records, nil
					},
				},
			},
		},
		{
			name:   "No URLs, status code: Unknown",
			userID: "1",
			want: want{
				code: codes.Unknown,
			},
			serv: services{
				provider: &providerMock{
					CollectByUserFn: func(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
						return []models.ShortenedURL{}, nil
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closer :=
				urlsProviderClient(context.Background(), tt.serv.provider)
			defer closer()

			reqCtx := metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("user_id", tt.userID))
			resp, err := client.ListURLs(reqCtx, &pb.ListURLsRequest{})
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.EqualValues(t, tt.want.code, e.Code(),
						"Expected status code: %d, got %d", tt.want.code, e.Code())
					return
				} else {
					t.Fatalf("failed to parse: %v", err)
				}
			}

			assert.EqualValues(t, len(tt.want.data.CollectedUrls), len(resp.CollectedUrls))
		})
	}
}

func urlsProviderClient(
	ctx context.Context,
	prov provider,
) (pb.URLsProviderClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterURLsProviderServer(
		baseServer,
		newURLsProviderService(prov),
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

	client := pb.NewURLsProviderClient(conn)
	return client, closer
}
