package http_mux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"simple-http-mux/internal/http_fetcher"
	"simple-http-mux/internal/models"
	"simple-http-mux/pkg/netutil"
	"time"
)

const (
	maxUrlsPerRequest = 20
)

type HttpMux struct {
	listener       net.Listener
	server         *http.Server
	port           uint16
	maxConnections uint
}

func NewHttpMux(port uint16, maxConnections uint) *HttpMux {

	router := http.NewServeMux()
	router.HandleFunc("/", handleRequest)

	server := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 1 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	return &HttpMux{
		server:         server,
		port:           port,
		maxConnections: maxConnections,
	}

}

func (h *HttpMux) Run() error {
	var err error
	if h.server == nil {
		return errors.New("HttpMux is not initialized, please use NewHttpMux()")
	}

	//create a default listener
	h.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		return err
	}

	//if we got max connections limitations, upgrade the default listener
	if h.maxConnections > 0 {
		h.listener = netutil.LimitListener(h.listener, int(h.maxConnections))
		log.Printf("Max connections set to %d...", h.maxConnections)
	}
	defer func() {
		err = h.listener.Close()
		if err != nil {
			log.Printf("Failed to close listener %s...", err.Error())
		}
		err = h.server.Shutdown(context.Background())
		if err != nil {
			log.Printf("Failed to close server %s...", err.Error())
		}
	}()
	log.Printf("HttpMux started. Listening on %s...", h.listener.Addr().String())

	err = h.server.Serve(h.listener)
	switch {
	case errors.Is(err, http.ErrServerClosed):
		log.Printf("HttpMux stopped gracefully.")
		return nil
	default:
		log.Printf("Error occured during HTTP server execution :%s", err.Error())
		return err
	}

}

func (h *HttpMux) Shutdown(ctx context.Context) error {
	err := h.server.Shutdown(ctx)
	return err
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming request from %s\n", r.RemoteAddr)
	//validate method (only POST)
	if r.Method != http.MethodPost {
		sendError(&w, "Only POST method is supported", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(&w, "Unable to read request body", http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		sendError(&w, "Request body is empty", http.StatusInternalServerError)
		return
	}
	var dto models.UrlsDto
	err = json.Unmarshal(body, &dto)
	if err != nil {
		sendError(&w, "Incorrect JSON in request body", http.StatusInternalServerError)
		return
	}
	if len(dto.Urls) > maxUrlsPerRequest {
		sendError(&w, "More then 20 urls in", http.StatusInternalServerError)
		return
	}

	go func(ctx context.Context) {
		resps, err := http_fetcher.FetchUrls(
			ctx,
			dto.Urls,
			2,
			10*time.Second,
			1*time.Second,
		)
		if err != nil {
			log.Printf("%v", err)
		}
		log.Printf("Resps len = %d", len(resps))
	}(context.Background())
	w.WriteHeader(http.StatusOK)
}
