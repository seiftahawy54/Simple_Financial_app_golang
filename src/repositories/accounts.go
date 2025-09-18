package repositories

import (
	"context"
	"errors"
	"finance_app/src/models"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccountsMongoRepository struct {
	collection *mongo.Collection
}

func (r *AccountsMongoRepository) FindOne(ctx context.Context, id string) (*models.Accounts, error) {
	if id == "" {
		return nil, errors.New("transaction ID cannot be empty")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID format: %w", err)
	}

	var account models.Accounts

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&account)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
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
