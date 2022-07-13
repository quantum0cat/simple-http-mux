package http_mux

import (
	"net/http"
)

//aux func to send error to writer
func sendError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
