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

type AccountsMongoRepository struct {
	collection *mongo.Collection
}

func NewAccountsMongoRepository(db *mongo.Database) *AccountsMongoRepository {
	return &AccountsMongoRepository{
		collection: db.Collection("accounts"),
	}
}

func (r *AccountsMongoRepository) FindOne(ctx context.Context, id string) (*models.Accounts, error) {
	if id == "" {
		return nil, errors.New("account ID cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, errors.New("invalid account ID format")
	}

	var account models.Accounts

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&account)

	if err != nil {
		return nil, errors.New("account not found")
	}

	return &account, nil
}

func (r *AccountsMongoRepository) UpdateBalance(ctx context.Context, id string, newBalance float64) error {
	account, err := r.FindOne(ctx, id)

	if err != nil {
		return err
	}

	if newBalance < 0 {
		return errors.New("balance cannot be negative")
	}

	account.Balance = newBalance

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": account.ID}, bson.M{"$set": bson.M{"balance": newBalance}})

	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

func (r *AccountsMongoRepository) GetAllAccounts(ctx context.Context) ([]models.Accounts, error) {
	var accounts []models.Accounts

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)

	if err != nil {
		return nil, errors.New("failed to fetch accounts")
	}

	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &accounts); err != nil {
		return nil, errors.New("failed to decode accounts")
	}

	return accounts, nil
}

func (r *AccountsMongoRepository) CreateAccount(ctx context.Context, account *models.Accounts) error {

	oldAccount := &models.Accounts{}

	err := r.collection.FindOne(ctx, bson.M{"email": account.Email}).Decode(&oldAccount)

	if err == nil && oldAccount != nil {
		account.ID = oldAccount.ID
		return errors.New("account with this email already exists")
	}

	account.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	account.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	// Insert into database
	result, err := r.collection.InsertOne(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Update transaction ID from result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		account.ID = oid
	}

	return nil
}
