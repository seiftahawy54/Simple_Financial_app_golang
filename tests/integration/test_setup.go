package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"finance_app/src/handlers"
	"finance_app/src/repositories"
	"finance_app/src/routes"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	MongoURI    string
	Database    string
	TestTimeout time.Duration
}

// TestSuite holds all test dependencies
type TestSuite struct {
	Client                *mongo.Client
	Database              *mongo.Database
	Handler               *handlers.AppHandler
	Router                chi.Router
	TransactionRepository *repositories.TransactionMongoRepository
	AccountsRepository    *repositories.AccountsMongoRepository
	Config                *TestConfig
}

// SetupTestSuite initializes the test environment
func SetupTestSuite(t *testing.T) *TestSuite {
	// Load test configuration
	config := &TestConfig{
		MongoURI:    getEnvOrDefault("TEST_MONGO_URI", "mongodb://localhost:27017"),
		Database:    getEnvOrDefault("TEST_DATABASE", "finance_test_db"),
		TestTimeout: 30 * time.Second,
	}

	// Connect to test MongoDB
	client, err := connectToTestMongo(config.MongoURI)
	if err != nil {
		t.Fatalf("Failed to connect to test MongoDB: %v", err)
	}

	// Get test database
	db := client.Database(config.Database)

	// Initialize repositories
	transactionRepo := repositories.NewTransactionMongoRepository(db)
	accountsRepo := repositories.NewAccountsMongoRepository(db)

	// Create handler with dependencies
	handler := handlers.NewAppHandler(client, *transactionRepo, *accountsRepo)

	// Setup router
	router := chi.NewRouter()
	routes.Routes(router, handler)

	return &TestSuite{
		Client:                client,
		Database:              db,
		Handler:               handler,
		Router:                router,
		TransactionRepository: transactionRepo,
		AccountsRepository:    accountsRepo,
		Config:                config,
	}
}

// CleanupTestSuite cleans up test data and closes connections
func (ts *TestSuite) CleanupTestSuite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Drop test database to clean up all data
	if err := ts.Database.Drop(ctx); err != nil {
		t.Logf("Warning: Failed to drop test database: %v", err)
	}

	// Disconnect from MongoDB
	if err := ts.Client.Disconnect(ctx); err != nil {
		t.Logf("Warning: Failed to disconnect from MongoDB: %v", err)
	}
}

// CleanupCollections cleans up specific collections
func (ts *TestSuite) CleanupCollections(t *testing.T, collections ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, collection := range collections {
		if err := ts.Database.Collection(collection).Drop(ctx); err != nil {
			t.Logf("Warning: Failed to drop collection %s: %v", collection, err)
		}
	}
}

// connectToTestMongo connects to test MongoDB instance
func connectToTestMongo(uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Verify the connection with a ping
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SkipIfNoMongo skips the test if MongoDB is not available
func SkipIfNoMongo(t *testing.T) {
	uri := getEnvOrDefault("TEST_MONGO_URI", "mongodb://localhost:27017")
	client, err := connectToTestMongo(uri)
	if err != nil {
		t.Skipf("Skipping test: MongoDB not available at %s: %v", uri, err)
	}

	// Clean up test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.Disconnect(ctx)
}
