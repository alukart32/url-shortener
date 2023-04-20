package v1

import (
	"context"
	"errors"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/services/shorturl"
	pb "github.com/alukart32/shortener-url/pkg/proto/v1/urls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// getterByURL defines the shortened URL provider by raw URL.
type getterByURL interface {
	GetByURL(context.Context, string) (models.ShortenedURL, error)
}

// shortener defines the url shortener.
type shortener interface {
	Short(context.Context, models.URL) (string, error)
	Batch(context.Context, []models.URL) ([]models.ShortenedURL, error)
}

// urlsShortenerService is a representation of the proto URLsShortenerServer.
type urlsShortenerService struct {
	pb.UnimplementedURLsShortenerServer
	provider  getterByURL
	shortener shortener
}

// newURLsShortenerService returns a new urlsShortenerService.
func newURLsShortenerService(
	provider getterByURL,
	shortener shortener,
) *urlsShortenerService {
	return &urlsShortenerService{
		provider:  provider,
		shortener: shortener,
	}
}

// ShortURL shortens the URL. If the URL has been shortened, its shortened form is returned.
func (s *urlsShortenerService) ShortURL(ctx context.Context, in *pb.ShortURLRequest) (*pb.ShortURLResponse, error) {
	var response pb.ShortURLResponse

	userID := getUserIDFromCtx(ctx)

	shortenedURL, err := s.shortener.Short(ctx, models.NewURL(userID, "", in.Raw))
	if err != nil {
		if errors.Is(err, shorturl.ErrUniqueViolation) {
			// Get an existing shortened URL.
			shortURL, err := s.provider.GetByURL(ctx, in.Raw)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			return nil, status.Errorf(codes.AlreadyExists, `found shortened URL: %s`, shortURL.Value)
		}
		if errors.Is(err, shorturl.ErrInvalidCreation) {
			err = status.Errorf(codes.InvalidArgument, "failed to short URL: %v", err)
		} else {
			err = status.Error(codes.Internal, err.Error())
		}
		return nil, err
	}

	response.ShortUrl = shortenedURL
	return &response, nil
}

// BatchURLs shortens a list of URLs.
func (s *urlsShortenerService) BatchURLs(ctx context.Context, in *pb.BatchURLsRequest) (*pb.BatchURLsResponse, error) {
	var response pb.BatchURLsResponse

	userID := getUserIDFromCtx(ctx)

	urlsToBatch := make([]models.URL, len(in.Url))
	for i, v := range in.Url {
		urlsToBatch[i] = models.NewURL(userID, v.CorrId, v.Raw)
	}
	urls, err := s.shortener.Batch(ctx, urlsToBatch)
	if err != nil {
		if errors.Is(err, shorturl.ErrEmptyBatch) ||
			errors.Is(err, shorturl.ErrInvalidCreation) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	respData := make([]*pb.BatchURLsResponse_URL, len(urls))
	for i, u := range urls {
		respData[i] = &pb.BatchURLsResponse_URL{
			CorrId:   u.CorrID,
			ShortUrl: u.Value,
		}
	}

	response.BatchedUrls = respData
	return &response, nil
}
