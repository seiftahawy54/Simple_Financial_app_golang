package services

import (
	"encoding/json"
	"finance_app/src/models"
	"finance_app/src/repositories"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response structures for consistent API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Request structure for creating/updating transactions
type CreateTransactionRequest struct {
	TransactionType string  `json:"transactionType"`
	Amount          float64 `json:"amount"`
	Balance         float64 `json:"balance"`
	AccountId       string  `json:"accountId"`
}

type TransactionHandler struct {
	Repo repositories.TransactionMongoRepository
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Error("Failed to encode response: ", err)
	}
}

// GetAllTransactions handles GET /api/v1/transactions
func (h *TransactionHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	transactions, err := h.Repo.GetAllTransactions(ctx)
	if err != nil {
		logrus.Error("Failed to get transactions: ", err)
		sendJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch transactions",
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    transactions,
		Message: "Transactions fetched successfully",
	})
}

// CreateTransaction handles POST /api/v1/transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate and parse account ID
	accountObjID, err := primitive.ObjectIDFromHex(req.AccountId)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid account ID format",
		})
		return
	}

	// Create transaction model
	transaction := &models.Transaction{
		TransactionType: models.TransactionType(strings.ToUpper(req.TransactionType)),
		Amount:          req.Amount,
		Balance:         req.Balance,
		AccountId:       accountObjID,
	}

	// Save to database
	if err := h.Repo.Create(ctx, transaction); err != nil {
		logrus.Error("Failed to create transaction: ", err)
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	sendJSONResponse(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    transaction,
		Message: "Transaction created successfully",
	})
}

// GetTransactionByID handles GET /api/v1/transactions/{id}
func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transactionID := chi.URLParam(r, "id")

	if transactionID == "" {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Transaction ID is required",
		})
		return
	}

	transaction, err := h.Repo.GetByID(ctx, transactionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendJSONResponse(w, http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "Transaction not found",
			})
			return
		}
		logrus.Error("Failed to get transaction: ", err)
		sendJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch transaction",
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    transaction,
		Message: "Transaction fetched successfully",
	})
}

// GetTransactionsByAccountID handles GET /api/v1/transactions/account/{accountId}
func (h *TransactionHandler) GetTransactionsByAccountID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "accountId")

	if accountID == "" {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Account ID is required",
		})
		return
	}

	transactions, err := h.Repo.GetByAccountID(ctx, accountID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid account ID") {
			sendJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Invalid account ID format",
			})
			return
		}
		logrus.Error("Failed to get transactions for account: ", err)
		sendJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to fetch transactions",
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    transactions,
		Message: "Transactions fetched successfully",
	})
}
