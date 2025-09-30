package services

import (
	"encoding/json"
	"finance_app/src/models"
	"finance_app/src/repositories"
	"finance_app/src/utils"
	"finance_app/src/utils/types"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type CreateAccountRequest struct {
	Balance float64 `json:"initialBalance"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
}

type AccountHandler struct {
	TransactionsRepo repositories.TransactionMongoRepository
	AccountsRepo     repositories.AccountsMongoRepository
}

// GetAllAccounts handles GET /api/v1/accounts
func (h *AccountHandler) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := h.AccountsRepo.GetAllAccounts(ctx)
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
		Data:    accounts,
		Message: "Accounts fetched successfully",
	})
}

// CreateAccount handles POST /api/v1/accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Name == "" {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Name is required",
		})
		return
	}

	if req.Email == "" {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Email is required",
		})
		return
	}

	account := models.Accounts{
		Balance: req.Balance,
		Name:    req.Name,
		Email:   req.Email,
	}

	err := h.AccountsRepo.CreateAccount(ctx, &account)

	if err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   err.Error(),
			Data:    account,
		})
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, types.APIResponse{
		Success: true,
		Data:    account,
		Message: "Account created successfully",
	})
}

// GetAccountByID handles GET /api/v1/accounts/{id}
func (h *AccountHandler) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		utils.SendJSONResponse(w, http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error:   "Account ID is required",
		})
		return
	}

	account, err := h.AccountsRepo.FindOne(ctx, id)

	if err != nil {
		logrus.Error("Failed to get account: ", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error:   "Failed to fetch account",
		})
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, types.APIResponse{
		Success: true,
		Data:    account,
		Message: "Account fetched successfully",
	})
}
