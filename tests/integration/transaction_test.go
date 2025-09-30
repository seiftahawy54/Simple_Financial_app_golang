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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTransactionIntegration(t *testing.T) {
	// Skip if MongoDB is not available
	SkipIfNoMongo(t)

	// Setup test suite
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	// Helper function to create a test account
	createTestAccount := func(name, email string, balance float64) *models.Accounts {
		account := &models.Accounts{
			Name:    name,
			Email:   email,
			Balance: balance,
		}
		err := ts.AccountsRepository.CreateAccount(context.Background(), account)
		require.NoError(t, err)
		return account
	}

	t.Run("Create Deposit Transaction", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create test account
		account := createTestAccount("John Doe", "john@example.com", 1000.0)
		accountID := account.ID.Hex()

		// Test data
		transactionData := map[string]interface{}{
			"transactionType": "DEPOSIT",
			"amount":          500.0,
			"accountId":       accountID,
		}

		// Create request
		jsonData, err := json.Marshal(transactionData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
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
		assert.Equal(t, "Transaction created successfully", response.Message)

		// Verify response data structure
		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data, "transactions")
		assert.Contains(t, data, "account")

		// Verify account balance was updated
		accountData, ok := data["account"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1500.0, accountData["balance"]) // 1000 + 500

		// Verify transaction was created
		transactions, ok := data["transactions"].([]interface{})
		require.True(t, ok)
		assert.Len(t, transactions, 1)

		transaction, ok := transactions[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "DEPOSIT", transaction["transactionType"])
		assert.Equal(t, 500.0, transaction["amount"])
	})

	t.Run("Create Withdraw Transaction", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create test account
		account := createTestAccount("John Doe", "john@example.com", 1000.0)
		accountID := account.ID.Hex()

		// Test data
		transactionData := map[string]interface{}{
			"transactionType": "WITHDRAW",
			"amount":          300.0,
			"accountId":       accountID,
		}

		// Create request
		jsonData, err := json.Marshal(transactionData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
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

		// Verify account balance was updated
		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		accountData, ok := data["account"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 700.0, accountData["balance"]) // 1000 - 300
	})

	t.Run("Create Transaction with Invalid Account", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Test with non-existent account ID
		invalidAccountID := primitive.NewObjectID().Hex()
		transactionData := map[string]interface{}{
			"transactionType": "DEPOSIT",
			"amount":          500.0,
			"accountId":       invalidAccountID,
		}

		jsonData, err := json.Marshal(transactionData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "account not found")
	})

	t.Run("Create Transaction with Invalid Type", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create test account
		account := createTestAccount("John Doe", "john@example.com", 1000.0)
		accountID := account.ID.Hex()

		// Test with invalid transaction type
		transactionData := map[string]interface{}{
			"transactionType": "INVALID_TYPE",
			"amount":          500.0,
			"accountId":       accountID,
		}

		jsonData, err := json.Marshal(transactionData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Invalid transaction type")
	})

	t.Run("Get All Transactions", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create test account
		account := createTestAccount("John Doe", "john@example.com", 1000.0)

		// Create test transactions
		transactions := []*models.Transaction{
			{
				TransactionType: models.Deposit,
				Amount:          500.0,
				AccountId:       account.ID,
			},
			{
				TransactionType: models.Withdraw,
				Amount:          200.0,
				AccountId:       account.ID,
			},
		}

		for _, transaction := range transactions {
			err := ts.TransactionRepository.Create(context.Background(), transaction)
			require.NoError(t, err)
		}

		// Test GET /api/v1/transactions
		req := httptest.NewRequest("GET", "/api/v1/transactions", nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response types.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Transactions fetched successfully", response.Message)

		// Verify transactions data
		transactionsList, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, transactionsList, 2)
	})

	t.Run("Get Transaction by ID", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create test account and transaction
		account := createTestAccount("John Doe", "john@example.com", 1000.0)
		transaction := &models.Transaction{
			TransactionType: models.Deposit,
			Amount:          500.0,
			AccountId:       account.ID,
		}
		err := ts.TransactionRepository.Create(context.Background(), transaction)
		require.NoError(t, err)

		transactionID := transaction.ID.Hex()

		// Test GET /api/v1/transactions/{id}
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/transactions/%s", transactionID), nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Transaction fetched successfully", response.Message)

		// Verify transaction data
		transactionData, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "DEPOSIT", transactionData["transactionType"])
		assert.Equal(t, 500.0, transactionData["amount"])
	})

	t.Run("Get Transactions by Account ID", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create test account
		account := createTestAccount("John Doe", "john@example.com", 1000.0)
		accountID := account.ID.Hex()

		// Create test transactions
		transactions := []*models.Transaction{
			{
				TransactionType: models.Deposit,
				Amount:          500.0,
				AccountId:       account.ID,
			},
			{
				TransactionType: models.Withdraw,
				Amount:          200.0,
				AccountId:       account.ID,
			},
		}

		for _, transaction := range transactions {
			err := ts.TransactionRepository.Create(context.Background(), transaction)
			require.NoError(t, err)
		}

		// Test GET /api/v1/transactions/account/{accountId}
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/transactions/account/%s", accountID), nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response types.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Transactions fetched successfully", response.Message)

		// Verify transactions data
		transactionsList, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, transactionsList, 2)
	})

	t.Run("Get Transaction by Invalid ID", func(t *testing.T) {
		// Test with invalid ObjectID
		req := httptest.NewRequest("GET", "/api/v1/transactions/invalid-id", nil)
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response types.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Failed to fetch transaction")
	})

	t.Run("Create Transaction with Invalid Data", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Test with missing required fields
		invalidData := map[string]interface{}{
			"transactionType": "DEPOSIT",
			// Missing amount and accountId
		}

		jsonData, err := json.Marshal(invalidData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "account ID cannot be empty")
	})
}
