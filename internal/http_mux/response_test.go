package http_mux

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const errorMessage = "TestErrorMessage!!!"

func Test_sendError(t *testing.T) {

	tests := []struct {
		name       string
		message    string
		statusCode int
	}{
		{
			name:       "default",
			message:    errorMessage,
			statusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			sendError(w, tt.message, tt.statusCode)

			r := w.Result()
			assert.Equal(t, r.StatusCode, tt.statusCode, "status codes don't match")

			defer func() { _ = r.Body.Close() }()
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err, "failed to read response body")

			assert.Equal(t, tt.message+"\n", string(body), "messages don't match")

			fmt.Printf("%s", string(body))
		})
	}
}
