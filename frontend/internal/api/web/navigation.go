package web

import (
	"encoding/json"
	"net/http"
)

// htmx ignores response headers on 3xx. Non-JS requests retain the 303.
func navigate(w http.ResponseWriter, r *http.Request, path string) {
	if r.Header.Get("HX-Request") == "true" {
		location, _ := json.Marshal(struct {
			Path   string `json:"path"`
			Target string `json:"target"`
			Select string `json:"select"`
			Swap   string `json:"swap"`
		}{path, "#page", "#page", "outerHTML"})
		w.Header().Set("HX-Location", string(location))
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}
