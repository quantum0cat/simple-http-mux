package http_mux

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"simple-http-mux/pkg/netutil"
	"time"
)

type HttpMux struct {
	listener       net.Listener
	server         *http.Server
	port           uint16
	maxConnections uint
}

func NewHttpMux(ctx context.Context, port uint16, maxConnections uint) *HttpMux {

	handler := newMuxHandler(ctx)

	server := &http.Server{
		Handler:           handler,
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
	maxConnsStr := ""
	if h.maxConnections > 0 {
		h.listener = netutil.LimitListener(h.listener, int(h.maxConnections))
		maxConnsStr = fmt.Sprintf("Max connections = %d", h.maxConnections)
	}
	defer func() {
		err = h.server.Shutdown(context.Background())
		if err != nil {
			log.Printf("Failed to close server %s...", err.Error())
		}
	}()

	log.Printf("HttpMux started. Listening on %s. %s", h.listener.Addr().String(), maxConnsStr)

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
