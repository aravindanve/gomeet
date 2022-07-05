package resource

import (
	"time"
)

const (
	UserProviderGoogle UserProvider = "google"
)

type UserProvider string

type User struct {
	ID                 ResourceID   `json:"id" bson:"_id,omitempty"`
	Name               string       `json:"name" bson:"name"`
	ImageURL           *string      `json:"imageUrl" bson:"imageUrl"`
	Provider           UserProvider `json:"provider" bson:"provider"`
	ProviderResourceID string       `json:"providerResourceId" bson:"providerResourceId"`
	CreatedAt          time.Time    `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt" bson:"updatedAt"`
}
