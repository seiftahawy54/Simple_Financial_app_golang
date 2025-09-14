package repositories

import (
	"context"
	"errors"
	"finance_app/src/models"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TransactionMongoRepository struct {
	collection *mongo.Collection
}

func NewTransactionMongoRepository(db *mongo.Database) *TransactionMongoRepository {
	return &TransactionMongoRepository{
		collection: db.Collection("transactions"),
	}
}

func (r *TransactionMongoRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	// Validate required fields
	if transaction == nil {
		return errors.New("transaction cannot be nil")
	}

	if transaction.TransactionType == "" {
		return errors.New("transaction type is required")
	}

	if transaction.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if transaction.AccountId.IsZero() {
		return errors.New("account ID is required")
	}

	// Enforce enum validation
	switch transaction.TransactionType {
	case models.Deposit, models.Withdraw, models.Transfer:
		// Valid transaction type
	default:
		return fmt.Errorf("invalid transaction type: %s", transaction.TransactionType)
	}

	// Set creation timestamp
	if transaction.ID.IsZero() {
		transaction.ID = primitive.NewObjectID()
	}
	transaction.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	// Insert into database
	result, err := r.collection.InsertOne(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update transaction ID from result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		transaction.ID = oid
	}

	return nil
}

func (r *TransactionMongoRepository) GetAllTransactions(ctx context.Context) ([]models.Transaction, error) {
	var transactions []models.Transaction

	// Find all transactions, sorted by creation date (newest first)
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode all transactions
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}

	// Check for cursor errors
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	// Return empty slice instead of nil if no transactions found
	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return transactions, nil
}

func (r *TransactionMongoRepository) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	// Validate and parse ID
	if id == "" {
		return nil, errors.New("transaction ID cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID format: %w", err)
	}

	// Find transaction
	var transaction models.Transaction
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("transaction not found with ID: %s", id)
		}
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	return &transaction, nil
}

func (r *TransactionMongoRepository) GetByAccountID(ctx context.Context, accountID string) ([]*models.Transaction, error) {
	// Validate and parse account ID
	if accountID == "" {
		return nil, errors.New("account ID cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID format: %w", err)
	}

	// Find transactions for account, sorted by date (newest first)
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"accountId": objID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions for account: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode all transactions
	var transactions []*models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}

	// Return empty slice instead of nil if no transactions found
	if transactions == nil {
		transactions = []*models.Transaction{}
	}

	return transactions, nil
}
