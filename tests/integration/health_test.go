package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	// Skip if MongoDB is not available
	SkipIfNoMongo(t)

	// Setup test suite
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	t.Run("Health Check", func(t *testing.T) {
		// Test GET /api/v1/health
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response structure
		assert.Equal(t, "OK", response["message"])
		assert.Equal(t, "healthy", response["status"])
		assert.Contains(t, response, "timestamp")

		// Verify timestamp is valid
		timestampStr, ok := response["timestamp"].(string)
		require.True(t, ok)
		_, err = time.Parse(time.RFC3339, timestampStr)
		assert.NoError(t, err, "Timestamp should be valid RFC3339 format")
	})
}

