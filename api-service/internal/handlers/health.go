package handlers

import (
	"net/http"
	"time"
)

func healthLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "go-api",
		"status":  "alive",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}
