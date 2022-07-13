package http_mux

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"simple-http-mux/internal/models"
	"testing"
	"time"
)

func TestNewHttpMux(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := NewHttpMux(context.Background(), 10000, 100)
			assert.NotNil(t, mux, "constructor returned nil")
		})
	}
}

const (
	testServersCount            = 10
	testServerResponseFormatIdx = "Test server response %d"
)

func generateHandlerFunc(serverId int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = fmt.Fprintf(w, testServerResponseFormatIdx, serverId)
	}
}

const serverPort = 10000

func TestHttpMux_Run(t *testing.T) {

	testServers := make([]*httptest.Server, testServersCount)
	urls := make([]string, testServersCount)
	responses := make([]models.Response, testServersCount)
	requests := make([]*http.Request, testServersCount)
	portUrl := fmt.Sprintf("http://localhost:%d", serverPort)
	var err error
	for i := 0; i < testServersCount; i++ {
		testServers[i] = httptest.NewServer(http.HandlerFunc(generateHandlerFunc(i)))
		urls[i] = testServers[i].URL
		responses[i] = models.Response{
			Url:      urls[i],
			Response: fmt.Sprintf(testServerResponseFormatIdx, i),
		}

	}
	dto := models.UrlsDto{Urls: urls}
	data := dto.Marshal()
	for i := 0; i < testServersCount; i++ {
		requests[i], err = http.NewRequest(
			http.MethodPost,
			portUrl,
			bytes.NewBuffer(data),
		)
		assert.NoError(t, err, "failed to make a new request")
	}

	tests := []struct {
		name           string
		port           uint16
		maxConnections uint
		requestsCount  int
		timeAlive      time.Duration
		wantErr        bool
	}{
		{
			name:           "default",
			port:           10000,
			maxConnections: 10,
			requestsCount:  1,
			timeAlive:      1 * time.Second,
		},

		{
			name:           "check rejects by server",
			port:           10000,
			maxConnections: 1,
			requestsCount:  10,
			timeAlive:      10 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mux := NewHttpMux(context.Background(), tt.port, tt.maxConnections)
			go func() {
				time.Sleep(2 * time.Second)
				_ = mux.Shutdown(context.Background())
			}()

			go func() {
				time.Sleep(100 * time.Millisecond) //wait for server listen
				for i := 0; i < tt.requestsCount && i < len(requests); i++ {
					go func(i int) {
						client := http.Client{
							Transport: http.DefaultTransport,
							Timeout:   0,
						}
						resp, err := client.Do(requests[i])
						if err != nil {
							log.Printf("Err : %v", err)
						} else {
							log.Printf("%v", resp.StatusCode)
						}

					}(i)
				}
			}()
			err := mux.Run()
			assert.NoError(t, err, "server run got error")
		})
	}
}
