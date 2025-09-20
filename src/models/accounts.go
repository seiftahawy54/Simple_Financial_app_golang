package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Accounts struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Balance   float64            `bson:"balance" json:"balance"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	CreatedAt primitive.DateTime `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
