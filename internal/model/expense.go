package model

import "time"

type Expense struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	ExpenseDate string    `json:"expense_date"`
	BuktiKey    string    `json:"bukti_key"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}