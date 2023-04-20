package v1

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/stretchr/testify/assert"
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

func TestStat(t *testing.T) {
	type services struct {
		stat statProvider
	}
	type want struct {
		data models.Stat
		code int
	}
	tests := []struct {
		serv services
		name string
		want want
	}{
		{
			name: "Get statistics, status code: Ok",
			want: want{
				code: http.StatusOK,
				data: models.Stat{
					URLs:  7,
					Users: 2,
				},
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
			name: "Stat error, status code: InternalServerError",
			want: want{
				code: http.StatusInternalServerError,
				data: models.Stat{
					URLs:  7,
					Users: 2,
				},
			},
			serv: services{
				stat: &statProviderMock{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.GET("/api/internal/stats", stat(tt.serv.stat))

			w := httptest.NewRecorder()
			// Prepare request.
			req := textReq(t, "/api/internal/stats", http.MethodGet, bytes.NewBufferString(""))
			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want.code, resp.StatusCode,
				"Expected status code: %d, got %d", tt.want.code, resp.StatusCode)
		})
	}
}
