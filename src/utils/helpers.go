package utils

import (
	"encoding/json"
	"finance_app/src/utils/types"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Helper function to send JSON responses
func SendJSONResponse(w http.ResponseWriter, status int, response types.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Error("Failed to encode response: ", err)
	}
}
