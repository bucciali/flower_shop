package models

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
	IsAvailable bool    `json:"is_available"`
	CreatedAt   string  `json:"created_at"`
}
