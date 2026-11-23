package web

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

func NewVersionedHandler(next http.Handler, version string, catalogRequests prometheus.Counter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Bookshop-Version", version)
		if r.Method == http.MethodGet && r.URL.Path == "/books" && r.Header.Get("X-Bookshop-Lab") == "rollout" {
			catalogRequests.Inc()
		}
		next.ServeHTTP(w, r)
	})
}
