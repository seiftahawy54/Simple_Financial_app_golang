package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance_app/src/models"
	"finance_app/src/utils/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountIntegration(t *testing.T) {
	// Skip if MongoDB is not available
	SkipIfNoMongo(t)

	// Setup test suite
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	t.Run("Create Account", func(t *testing.T) {
		// Clean up accounts collection before test
		ts.CleanupCollections(t, "accounts")

		// Test data
		accountData := map[string]interface{}{
			"name":           "John Doe",
			"email":          "john.doe@example.com",
			"initialBalance": 1000.0,
		}

		// Create request
		jsonData, err := json.Marshal(accountData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Account created successfully", response.Message)

		// Verify account data
		account, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John Doe", account["name"])
		assert.Equal(t, "john.doe@example.com", account["email"])
		assert.Equal(t, 1000.0, account["balance"])
		assert.NotEmpty(t, account["id"])
	})

	t.Run("Create Account with Duplicate Email", func(t *testing.T) {
		// Clean up accounts collection before test
		ts.CleanupCollections(t, "accounts")

		// Create first account
		accountData := map[string]interface{}{
			"name":           "John Doe",
			"email":          "john.doe@example.com",
			"initialBalance": 1000.0,
		}

		jsonData, err := json.Marshal(accountData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Try to create account with same email
		req2 := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(jsonData))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()

		ts.Router.ServeHTTP(w2, req2)

		// Should return error
		assert.Equal(t, http.StatusBadRequest, w2.Code)

		var response types.APIResponse
		err = json.Unmarshal(w2.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "account with this email already exists")
	})

	t.Run("Get All Accounts", func(t *testing.T) {
		// Clean up accounts collection before test
		ts.CleanupCollections(t, "accounts")

		// Create test accounts
		accounts := []map[string]interface{}{
			{"name": "John Doe", "email": "john@example.com", "initialBalance": 1000.0},
			{"name": "Jane Smith", "email": "jane@example.com", "initialBalance": 2000.0},
		}

		// Insert accounts directly into database
		for _, accountData := range accounts {
			account := &models.Accounts{
				Name:    accountData["name"].(string),
				Email:   accountData["email"].(string),
				Balance: accountData["initialBalance"].(float64),
			}
			err := ts.AccountsRepository.CreateAccount(context.Background(), account)
			require.NoError(t, err)
		}

		// Test GET /api/v1/accounts
		req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response types.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Accounts fetched successfully", response.Message)

		// Verify accounts data
		accountsList, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, accountsList, 2)
	})

	t.Run("Get Account by ID", func(t *testing.T) {
		// Clean up accounts collection before test
		ts.CleanupCollections(t, "accounts")

		// Create test account
		account := &models.Accounts{
			Name:    "John Doe",
			Email:   "john@example.com",
			Balance: 1000.0,
		}
		err := ts.AccountsRepository.CreateAccount(context.Background(), account)
		require.NoError(t, err)

		accountID := account.ID.Hex()

		// Test GET /api/v1/accounts/{id}
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%s", accountID), nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Account fetched successfully", response.Message)

		// Verify account data
		accountData, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John Doe", accountData["name"])
		assert.Equal(t, "john@example.com", accountData["email"])
		assert.Equal(t, 1000.0, accountData["balance"])
	})

	t.Run("Get Account by Invalid ID", func(t *testing.T) {
		// Test with invalid ObjectID
		req := httptest.NewRequest("GET", "/api/v1/accounts/invalid-id", nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response types.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Failed to fetch account")
	})

	t.Run("Create Account with Invalid Data", func(t *testing.T) {
		// Clean up accounts collection before test
		ts.CleanupCollections(t, "accounts")

		// Test with missing required fields
		invalidData := map[string]interface{}{
			"name": "John Doe",
			// Missing email and initialBalance
		}

		jsonData, err := json.Marshal(invalidData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Email is required")
	})
}
