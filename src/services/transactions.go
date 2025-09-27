package services

import (
	"encoding/json"
	"finance_app/src/models"
	"finance_app/src/repositories"
	"finance_app/src/utils"
	"finance_app/src/utils/types"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// Request structure for creating/updating transactions
type CreateTransactionRequest struct {
	TransactionType string  `json:"transactionType"`
	Amount          float64 `json:"amount"`
	AccountId       string  `json:"accountId"`
}

type TransactionHandler struct {
	TransactionsRepo repositories.TransactionMongoRepository
	AccountsRepo     repositories.AccountsMongoRepository
}

// GetAllTransactions handles GET /api/v1/transactions
func (h *TransactionHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	transactions, err := h.TransactionsRepo.GetAllTransactions(ctx)
	if err != nil {
		logrus.Error("Failed to get transactions: ", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   "Failed to fetch transactions",
		})
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, types.APIResponse{
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
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Validate and parse account ID
	account, err := h.AccountsRepo.FindOne(ctx, req.AccountId)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	accountId := account.ID.Hex()
	transactionType := models.TransactionType(strings.ToUpper(req.TransactionType))
	switch transactionType {
	case models.Deposit:
		account.Balance += req.Amount
	case models.Withdraw:
		account.Balance -= req.Amount
	default:
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Invalid transaction type",
		})
	}

	err = h.AccountsRepo.UpdateBalance(ctx, accountId, account.Balance)

	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Create transaction model
	transaction := &models.Transaction{
		TransactionType: models.TransactionType(strings.ToUpper(req.TransactionType)),
		Amount:          req.Amount,
		AccountId:       account.ID,
	}

	// Save to database
	if err := h.TransactionsRepo.Create(ctx, transaction); err != nil {
		logrus.Error("Failed to create transaction: ", err)
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	updatedAccount, err := h.AccountsRepo.FindOne(ctx, accountId)

	if err != nil {
		logrus.Error("Failed to fetch updated account: ", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   "Failed to fetch updated account",
		})
		return
	}

	transactionsForTheUser, err := h.TransactionsRepo.GetByAccountID(ctx, accountId)

	if err != nil {
		logrus.Error("Failed to fetch transactions for the user: ", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   "Failed to fetch transactions for the user",
		})
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, types.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"transactions": transactionsForTheUser,
			"account":      updatedAccount,
		},
		Message: "Transaction created successfully",
	})
}

// GetTransactionByID handles GET /api/v1/transactions/{id}
func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transactionID := chi.URLParam(r, "id")

	if transactionID == "" {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Transaction ID is required",
		})
		return
	}

	transaction, err := h.TransactionsRepo.GetByID(ctx, transactionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.SendJSONResponse(w, http.StatusNotFound, types.APIResponse{
				Success: false,
				Error:   "Transaction not found",
			})
			return
		}
		logrus.Error("Failed to get transaction: ", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   "Failed to fetch transaction",
		})
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, types.APIResponse{
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
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Account ID is required",
		})
		return
	}

	transactions, err := h.TransactionsRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid account ID") {
			utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
				Success: false,
				Error:   "Invalid account ID format",
			})
			return
		}
		logrus.Error("Failed to get transactions for account: ", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   "Failed to fetch transactions",
		})
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, types.APIResponse{
		Success: true,
		Data:    transactions,
		Message: "Transactions fetched successfully",
	})
}
