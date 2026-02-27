package model

import "time"

type Alert struct {
	ID             string    `json:"id" firestore:"id"`
	Title          string    `json:"title" firestore:"title"`
	Details        string    `json:"details" firestore:"details"`
	CreatedAt      time.Time `json:"created_at" firestore:"created_at"`
	IsAcknowledged bool      `json:"is_acknowledged" firestore:"is_acknowledged"`
}

type OnCall struct {
	Name        string
	PhoneNumber string
}
