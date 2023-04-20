package v1

import (
	"context"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	pb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// provider defines the shortened URL provider.
type provider interface {
	GetBySlug(context.Context, string) (models.ShortenedURL, error)
	CollectByUser(context.Context, string) ([]models.ShortenedURL, error)
}

// urlsShortenerService is a representation of the proto URLsProviderServer.
type urlsProviderService struct {
	pb.UnimplementedURLsProviderServer
	provider provider
}

// newURLsProviderService returns a new urlsProviderService.
func newURLsProviderService(provider provider) *urlsProviderService {
	return &urlsProviderService{
		provider: provider,
	}
}

// GetShortenedURL gets a shortened URL by the slug value.
func (s *urlsProviderService) GetShortenedURL(ctx context.Context, in *pb.GetShortenedURLRequest) (*pb.GetShortenedURLResponse, error) {
	if len(in.Slug) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "empty slug")
	}

	shortenedURL, err := s.provider.GetBySlug(ctx, in.Slug)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if shortenedURL.Empty() {
		return nil, status.Errorf(codes.NotFound, "shortened URL not found by %v", in.Slug)
	}
	if shortenedURL.IsDeleted {
		return nil, status.Error(codes.Unknown, "shortened URL was deleted")
	}

	var response pb.GetShortenedURLResponse
	response.ShortUrl = &pb.GetShortenedURLResponse_ShortenedURL{
		UserId:    shortenedURL.UserID,
		CorrId:    shortenedURL.CorrID,
		Raw:       shortenedURL.Raw,
		Slug:      shortenedURL.Slug,
		Value:     shortenedURL.Value,
		IsDeleted: shortenedURL.IsDeleted,
	}
	return &response, nil
}

// ListURLs collects user's shortened URLs.
func (s *urlsProviderService) ListURLs(ctx context.Context, in *pb.ListURLsRequest) (*pb.ListURLsResponse, error) {
	userID := getUserIDFromCtx(ctx)

	urls, err := s.provider.CollectByUser(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(urls) == 0 {
		return nil, status.Errorf(codes.Unknown, "no content")
	}

	list := make([]*pb.ListURLsResponse_URL, len(urls))
	for i, u := range urls {
		list[i] = &pb.ListURLsResponse_URL{
			Raw:      u.Raw,
			ShortUrl: u.Value,
		}
	}

	var response pb.ListURLsResponse
	response.CollectedUrls = list
	return &response, nil
}
