package handlers

import (
	"finance_app/src/repositories"
	"finance_app/src/services"
	"go.mongodb.org/mongo-driver/mongo"
)

// AppHandler holds dependencies like the DB client and services
type AppHandler struct {
	TransactionRepository repositories.TransactionMongoRepository
	TransactionService    *services.TransactionHandler
	Client                *mongo.Client
}

// NewAppHandler creates a new AppHandler with initialized services
func NewAppHandler(client *mongo.Client, transactionRepo repositories.TransactionMongoRepository) *AppHandler {
	transactionService := &services.TransactionHandler{
		Repo: transactionRepo,
	}

	return &AppHandler{
		TransactionRepository: transactionRepo,
		TransactionService:    transactionService,
		Client:                client,
	}
}
