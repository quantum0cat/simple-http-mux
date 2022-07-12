package http_mux

import (
	"fmt"
	"net/http"
)

func sendError(w *http.ResponseWriter, message string, statusCode int) {
	http.Error(*w, fmt.Sprintf(`{"error":"%s"}`, message), statusCode)
}
