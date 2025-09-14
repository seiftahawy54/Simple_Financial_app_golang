package main

import (
	"context"
	"finance_app/src/repositories"
	"net/http"
	"time"

	"finance_app/src/handlers"
	"finance_app/src/routes"
	"finance_app/src/utils"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Info("Starting server")

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		logrus.Warn("Warning: .env file not found, using system environment variables")
	}

	// Connect to MongoDB
	logrus.Info("Connecting to MongoDB...")
	client, err := utils.MongoConnect()
	if err != nil {
		logrus.Fatal("Failed to connect to MongoDB: ", err)
		// logrus.Fatal already calls os.Exit(1), so this return is unreachable but makes intent clear
		return
	}
	logrus.Info("Successfully connected to MongoDB")

	// Ensure MongoDB disconnection on exit
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logrus.Error("Error disconnecting from MongoDB: ", err)
		}
	}()

	// Initialize database and repository
	db := client.Database("finance_db")
	transactionRepo := repositories.NewTransactionMongoRepository(db)

	// Create handler with dependencies
	h := handlers.NewAppHandler(client, *transactionRepo)

	// Setup router
	router := chi.NewRouter()
	routes.Routes(router, h)

	// Start server
	logrus.Info("Server starting on port 1234")
	if err := http.ListenAndServe(":1234", router); err != nil {
		logrus.Fatal("Failed to start server: ", err)
	}
}
