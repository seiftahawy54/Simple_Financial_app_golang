package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance_app/src/utils/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndWorkflow(t *testing.T) {
	// Skip if MongoDB is not available
	SkipIfNoMongo(t)

	// Setup test suite
	ts := SetupTestSuite(t)
	defer ts.CleanupTestSuite(t)

	// Clean up collections before test
	ts.CleanupCollections(t, "accounts", "transactions")

	t.Run("Complete User Journey", func(t *testing.T) {
		// Step 1: Create an account
		t.Log("Step 1: Creating account...")
		accountData := map[string]interface{}{
			"name":           "Alice Johnson",
			"email":          "alice.johnson@example.com",
			"initialBalance": 1000.0,
		}

		jsonData, err := json.Marshal(accountData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var accountResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &accountResponse)
		require.NoError(t, err)
		assert.True(t, accountResponse.Success)

		account, ok := accountResponse.Data.(map[string]interface{})
		require.True(t, ok)
		accountID := account["id"].(string)
		assert.Equal(t, 1000.0, account["balance"])

		// Step 2: Make a deposit
		t.Log("Step 2: Making a deposit...")
		depositData := map[string]interface{}{
			"transactionType": "DEPOSIT",
			"amount":          500.0,
			"accountId":       accountID,
		}

		jsonData, err = json.Marshal(depositData)
		require.NoError(t, err)

		req = httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var depositResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &depositResponse)
		require.NoError(t, err)
		assert.True(t, depositResponse.Success)

		// Verify account balance after deposit
		data, ok := depositResponse.Data.(map[string]interface{})
		require.True(t, ok)
		updatedAccount, ok := data["account"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1500.0, updatedAccount["balance"]) // 1000 + 500

		// Step 3: Make a withdrawal
		t.Log("Step 3: Making a withdrawal...")
		withdrawData := map[string]interface{}{
			"transactionType": "WITHDRAW",
			"amount":          200.0,
			"accountId":       accountID,
		}

		jsonData, err = json.Marshal(withdrawData)
		require.NoError(t, err)

		req = httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var withdrawResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &withdrawResponse)
		require.NoError(t, err)
		assert.True(t, withdrawResponse.Success)

		// Verify account balance after withdrawal
		data, ok = withdrawResponse.Data.(map[string]interface{})
		require.True(t, ok)
		updatedAccount, ok = data["account"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1300.0, updatedAccount["balance"]) // 1500 - 200

		// Step 4: Get account details
		t.Log("Step 4: Getting account details...")
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%s", accountID), nil)
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var getAccountResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &getAccountResponse)
		require.NoError(t, err)
		assert.True(t, getAccountResponse.Success)

		accountDetails, ok := getAccountResponse.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Alice Johnson", accountDetails["name"])
		assert.Equal(t, "alice.johnson@example.com", accountDetails["email"])
		assert.Equal(t, 1300.0, accountDetails["balance"])

		// Step 5: Get all transactions for the account
		t.Log("Step 5: Getting account transactions...")
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/transactions/account/%s", accountID), nil)
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var transactionsResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &transactionsResponse)
		require.NoError(t, err)
		assert.True(t, transactionsResponse.Success)

		transactions, ok := transactionsResponse.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, transactions, 2) // Deposit and withdrawal

		// Verify transaction order (newest first)
		firstTransaction, ok := transactions[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "WITHDRAW", firstTransaction["transactionType"])
		assert.Equal(t, 200.0, firstTransaction["amount"])

		secondTransaction, ok := transactions[1].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "DEPOSIT", secondTransaction["transactionType"])
		assert.Equal(t, 500.0, secondTransaction["amount"])

		// Step 6: Get all accounts
		t.Log("Step 6: Getting all accounts...")
		req = httptest.NewRequest("GET", "/api/v1/accounts", nil)
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var allAccountsResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &allAccountsResponse)
		require.NoError(t, err)
		assert.True(t, allAccountsResponse.Success)

		allAccounts, ok := allAccountsResponse.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, allAccounts, 1)

		// Step 7: Get all transactions
		t.Log("Step 7: Getting all transactions...")
		req = httptest.NewRequest("GET", "/api/v1/transactions", nil)
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var allTransactionsResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &allTransactionsResponse)
		require.NoError(t, err)
		assert.True(t, allTransactionsResponse.Success)

		allTransactions, ok := allTransactionsResponse.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, allTransactions, 2)

		t.Log("End-to-end workflow completed successfully!")
	})

	t.Run("Multiple Users Workflow", func(t *testing.T) {
		// Clean up collections before test
		ts.CleanupCollections(t, "accounts", "transactions")

		// Create two accounts
		accounts := []map[string]interface{}{
			{"name": "User 1", "email": "user1@example.com", "initialBalance": 1000.0},
			{"name": "User 2", "email": "user2@example.com", "initialBalance": 2000.0},
		}

		var accountIDs []string
		for _, accountData := range accounts {
			jsonData, err := json.Marshal(accountData)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ts.Router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code)

			var response types.APIResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.True(t, response.Success)

			account, ok := response.Data.(map[string]interface{})
			require.True(t, ok)
			accountIDs = append(accountIDs, account["id"].(string))
		}

		// User 1 makes a deposit
		depositData := map[string]interface{}{
			"transactionType": "DEPOSIT",
			"amount":          300.0,
			"accountId":       accountIDs[0],
		}

		jsonData, err := json.Marshal(depositData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// User 2 makes a withdrawal
		withdrawData := map[string]interface{}{
			"transactionType": "WITHDRAW",
			"amount":          500.0,
			"accountId":       accountIDs[1],
		}

		jsonData, err = json.Marshal(withdrawData)
		require.NoError(t, err)

		req = httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify final balances
		req = httptest.NewRequest("GET", "/api/v1/accounts", nil)
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		allAccounts, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, allAccounts, 2)

		// Verify all transactions exist
		req = httptest.NewRequest("GET", "/api/v1/transactions", nil)
		w = httptest.NewRecorder()

		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var transactionsResponse types.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &transactionsResponse)
		require.NoError(t, err)
		assert.True(t, transactionsResponse.Success)

		allTransactions, ok := transactionsResponse.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, allTransactions, 2)

		t.Log("Multiple users workflow completed successfully!")
	})
}
