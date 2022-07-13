package http_mux

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"simple-http-mux/internal/http_fetcher"
	"simple-http-mux/internal/models"
	"simple-http-mux/pkg/utils"
	"sync/atomic"
	"time"
)

const maxUrlsPerRequest = 20

type muxHandler struct {
	ctx context.Context
	rid uint32
}

func newMuxHandler(ctx context.Context) *muxHandler {
	return &muxHandler{
		ctx: ctx,
		rid: 0,
	}
}

func (h *muxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rid := atomic.AddUint32(&h.rid, 1)

	log.Printf("Incoming request from %s\n", r.RemoteAddr)
	//validate method (only POST)
	if r.Method != http.MethodPost {
		sendError(w, utils.WithRid("Only POST method is supported", rid), http.StatusMethodNotAllowed)
		return
	}
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, utils.WithRid("Unable to read request body", rid), http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		sendError(w, utils.WithRid("Request body is empty", rid), http.StatusInternalServerError)
		return
	}
	var dto models.UrlsDto
	err = json.Unmarshal(body, &dto)
	if err != nil {
		sendError(w, utils.WithRid("Incorrect JSON in request body", rid), http.StatusInternalServerError)
		return
	}
	if len(dto.Urls) > maxUrlsPerRequest {
		sendError(w, utils.WithRid("More then 20 urls in", rid), http.StatusInternalServerError)
		return
	}

	fetcher, err := http_fetcher.NewHttpFetcher(
		rid,
		dto.Urls,
		4,
		10*time.Second,
		1*time.Second,
	)
	if err != nil {
		sendError(w, utils.WithRid(err.Error(), rid), http.StatusInternalServerError)
		return
	}

	resps, err := fetcher.Fetch(h.ctx)
	if err != nil {
		log.Printf("%s", utils.WithRid(err.Error(), rid))
		sendError(w, utils.WithRid(err.Error(), rid), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resps)
	if err != nil {
		sendError(w, utils.WithRid(err.Error(), rid), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Failed to write data to response : %s", utils.WithRid(err.Error(), rid))
	}
}
