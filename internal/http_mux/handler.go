package http_mux

import (
	"context"
	"fmt"
	"net/http"
)

type Handler struct {
	ctx context.Context
}

func NewHandler(ctx context.Context) *Handler {
	return &Handler{ctx: ctx}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}
