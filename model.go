package main

import "time"

// Email represents the structure of an email in the system.
type Email struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Status string `json:"status"`
	SentAt time.Time `json:"sent_at"`
	Details *string `json:"details"`
}