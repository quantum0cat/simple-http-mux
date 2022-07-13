package http_fetcher

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"simple-http-mux/internal/models"
	"testing"
	"time"
)

const (
	testServerResponseFormat = "Test server response"
)

func TestHttpFetcher_fetchUrl(t *testing.T) {

	fetcher := HttpFetcher{}
	testServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, testServerResponseFormat)
		},
	))

	testServerWithTimeout := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1500 * time.Millisecond)
			_, _ = fmt.Fprintf(w, testServerResponseFormat)
		},
	))

	var tests = []struct {
		name    string
		ctx     context.Context
		client  *http.Client
		url     string
		want    *models.Response
		wantErr bool
	}{
		{
			name: "without timeout",
			ctx:  context.Background(),
			client: &http.Client{
				Transport: http.DefaultTransport,
				Timeout:   0,
			},
			url: testServer.URL,
			want: &models.Response{
				Url:      testServer.URL,
				Response: "Test server response",
			},
			wantErr: false,
		},
		{
			name: "with timeout",
			ctx:  context.Background(),
			client: &http.Client{
				Transport: http.DefaultTransport,
				Timeout:   1 * time.Second,
			},
			url: testServerWithTimeout.URL,
			want: &models.Response{
				Url:      testServerWithTimeout.URL,
				Response: testServerResponseFormat,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := fetcher.fetchUrl(tt.ctx, tt.client, tt.url)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("fetchUrl() error = %v", err)
				return
			}
			assert.Equal(t, resp, tt.want, "responses not equal")
		})
	}
}

const (
	testServersCount            = 9
	testServerResponseFormatIdx = "Test server response %d"
)

func generateHandlerFunc(serverId int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = fmt.Fprintf(w, testServerResponseFormatIdx, serverId)
	}
}
func TestHttpFetcher_Fetch(t *testing.T) {

	testServers := make([]*httptest.Server, testServersCount)
	urls := make([]string, testServersCount)
	responses := make([]models.Response, testServersCount)
	for i := 0; i < testServersCount; i++ {
		testServers[i] = httptest.NewServer(http.HandlerFunc(generateHandlerFunc(i)))
		urls[i] = testServers[i].URL
		responses[i] = models.Response{
			Url:      urls[i],
			Response: fmt.Sprintf(testServerResponseFormatIdx, i),
		}
	}

	fetcher, err := NewHttpFetcher(
		0,
		urls,
		4,
		10*time.Second,
		1*time.Second,
	)
	assert.NoError(t, err, "failed to construct HttpFetcher")

	ctxC, cancel := context.WithCancel(context.Background())
	defer cancel()
	tests := []struct {
		name      string
		ctx       context.Context
		cancel    context.CancelFunc
		want      []models.Response
		wantError bool
	}{
		{
			name:   "10 servers fetch",
			ctx:    context.Background(),
			cancel: nil,
			want:   responses,
		},
		{
			name:      "10 servers fetch with cancelation",
			ctx:       ctxC,
			cancel:    cancel,
			want:      responses,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cancel != nil {
				go func() {
					time.Sleep(50 * time.Millisecond)
					tt.cancel()
				}()
			}
			resps, err := fetcher.Fetch(tt.ctx)
			if !tt.wantError {
				assert.NoError(t, err, "finished with error")
				assert.ElementsMatch(t, resps, tt.want, "responses don't match")
			}

		})
	}
}

func TestNewHttpFetcher(t *testing.T) {

	tests := []struct {
		name           string
		rid            uint32
		urls           []string
		maxWorkers     int
		fetchTimeout   time.Duration
		requestTimeout time.Duration
		want           *HttpFetcher
		wantErr        bool
	}{
		{
			name:           "default",
			rid:            0,
			urls:           []string{"url1", "url2"},
			maxWorkers:     1,
			fetchTimeout:   1,
			requestTimeout: 1,
			wantErr:        false,
		},
		{
			name:           "failing",
			rid:            0,
			urls:           []string{},
			maxWorkers:     1,
			fetchTimeout:   1,
			requestTimeout: 1,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := NewHttpFetcher(
				tt.rid,
				tt.urls,
				tt.maxWorkers,
				tt.fetchTimeout,
				tt.requestTimeout,
			)
			if !tt.wantErr {
				assert.NoError(t, err, "constructor returned error")
			}
		})
	}
}
