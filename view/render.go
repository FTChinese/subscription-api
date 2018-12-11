package view

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Render responds to client request
func Render(w http.ResponseWriter, resp Response) {
	// Set response headers
	for key, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(key, v)
		}
	}

	// If `Content-Type` is not set, set the json
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	// If there's no content body, or status code is 204, stop here.
	if resp.Body == nil || resp.StatusCode == http.StatusNoContent {
		w.WriteHeader(resp.StatusCode)
		return
	}

	w.WriteHeader(resp.StatusCode)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")

	// Write data to w
	err := enc.Encode(resp.Body)

	if err != nil {
		log.Printf("Render response error: %v\n", err)
		return
	}
}
