package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// TransactionTypes enums (adapt from your Object.values(TransactionsTypes))
type TransactionType string

const (
	Deposit  TransactionType = "DEPOSIT"
	Withdraw TransactionType = "WITHDRAW"
	Transfer TransactionType = "TRANSFER"
	// Add more as needed
)

// Transaction model based on schema
type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TransactionType TransactionType    `bson:"transactionType" json:"transactionType"`
	Amount          float64            `bson:"amount" json:"amount"`
	Balance         float64            `bson:"balance" json:"balance"`
	AccountId       primitive.ObjectID `bson:"accountId" json:"accountId"`
	CreatedAt       primitive.DateTime `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt       primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
