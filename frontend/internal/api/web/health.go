package web

import "net/http"

func serveHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
