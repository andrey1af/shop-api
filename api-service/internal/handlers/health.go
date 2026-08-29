package handlers

import (
	"net/http"
)

func healthLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func healthCheck(checker readinessChecker, unavailableStatus int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if checker == nil || checker.Ping(r.Context()) != nil {
			writeJSON(w, unavailableStatus, map[string]string{"status": "unavailable"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
