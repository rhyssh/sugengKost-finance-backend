package model

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Amount       float64   `json:"amount"`
	MonthYear    string    `json:"month_year"` // "2025-02"
	Status       string    `json:"status"`     // pending, verified, rejected
	BuktiURL     string    `json:"bukti_url"`
	Notes        string    `json:"notes,omitempty"`
	PaymentDate  time.Time `json:"payment_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}