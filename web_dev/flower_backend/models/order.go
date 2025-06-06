package models

import "time"

// Order — структура, описывающая заказ в таблице orders.
type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ProductIDs []int     `json:"product_ids"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
