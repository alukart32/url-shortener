package v1

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestDeleteURLsRoute_Delete(t *testing.T) {
	type services struct {
		deleter deleter
	}
	type want struct {
		code int
	}
	tests := []struct {
		serv services
		name string
		req  string
		want want
	}{
		{
			name: "Delete URLs, status code: Accepted",
			req:  `[ "slug1", "slug2", "slug3"]`,
			want: want{
				code: http.StatusAccepted,
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
			name: "No slugs for deleting, status code: NoContent",
			req:  `[]`,
			want: want{
				code: http.StatusNoContent,
			},
		},
		{
			name: "Delete error, status code: InternalServerError",
			req:  `[ "slug1", "slug2", "slug3"]`,
			want: want{
				code: http.StatusInternalServerError,
			},
			serv: services{
				deleter: &deleterMock{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin.Router.
			r := setupGin()
			r.DELETE("/api/user/urls", authWrap(deleteURLs(tt.serv.deleter)))

			w := httptest.NewRecorder()
			// Prepare the request.
			req := textReq(t, "/api/user/urls", http.MethodDelete, bytes.NewBufferString(tt.req))
			r.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.EqualValues(t, tt.want.code, resp.StatusCode)
		})
	}
}
