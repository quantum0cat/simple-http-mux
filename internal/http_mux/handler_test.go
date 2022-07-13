package http_mux

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"simple-http-mux/internal/models"
	"testing"
)

func Test_muxHandler_ServeHTTP(t *testing.T) {

	handler := &muxHandler{
		ctx: context.Background(),
	}

	//generate not allowed count of urls
	urls := make([]string, maxUrlsPerRequest+1)

	for i := 0; i <= maxUrlsPerRequest; i++ {
		urls[i] = fmt.Sprintf("url%d", i)
	}
	dto := models.UrlsDto{Urls: urls}
	unallowedUrlsData, err := json.Marshal(dto)
	fmt.Printf("%v", string(unallowedUrlsData))

	assert.NoError(t, err, "failed to make json from urls")

	tests := []struct {
		name       string
		request    *http.Request
		statusCode int
		errMessage string
	}{
		{
			name:       "failing1",
			request:    httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "failing2",
			request:    httptest.NewRequest(http.MethodPost, "http://localhost", nil),
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "failing3",
			request: httptest.NewRequest(
				http.MethodPost,
				"http://localhost",
				bytes.NewBuffer([]byte(`[`)),
			),
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "failing4",
			request: httptest.NewRequest(
				http.MethodPost,
				"http://localhost",
				bytes.NewBuffer(unallowedUrlsData),
			),
			statusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tt.request)
			r := w.Result()
			assert.Equal(t, r.StatusCode, tt.statusCode, "status codes don't match")

			defer func() { _ = r.Body.Close() }()
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err, "failed to read response body")

			fmt.Printf("%s", string(body))

		})
	}
}
