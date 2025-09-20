package handlers

import (
	"finance_app/src/repositories"
	"finance_app/src/services"

	"go.mongodb.org/mongo-driver/mongo"
)

// AppHandler holds dependencies like the DB client and services
type AppHandler struct {
	TransactionRepository repositories.TransactionMongoRepository
	AccountsRepository    repositories.AccountsMongoRepository
	TransactionService    *services.TransactionHandler
	AccountService        *services.AccountHandler
	Client                *mongo.Client
}

// NewAppHandler creates a new AppHandler with initialized services
func NewAppHandler(client *mongo.Client, transactionRepo repositories.TransactionMongoRepository, accountsRepo repositories.AccountsMongoRepository) *AppHandler {
	transactionService := &services.TransactionHandler{
		TransactionsRepo: transactionRepo,
		AccountsRepo:     accountsRepo,
	}

	accountService := &services.AccountHandler{
		AccountsRepo:     accountsRepo,
		TransactionsRepo: transactionRepo,
	}

	return &AppHandler{
		TransactionRepository: transactionRepo,
		AccountsRepository:    accountsRepo,
		TransactionService:    transactionService,
		AccountService:        accountService,
		Client:                client,
	}
}
