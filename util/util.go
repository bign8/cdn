package util

import "net/http"

// Static creates a static http response handler
func Static(msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(msg))
	}
}
