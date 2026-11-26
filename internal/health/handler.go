package health

import (
	"encoding/json"
	"net/http"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	payload := map[string]string{
		"status":    http.StatusText(http.StatusOK),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(payload)
}
