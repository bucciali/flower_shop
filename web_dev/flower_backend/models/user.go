package models

type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"`
	NewPassword string `json:"new_password,omitempty"` // для смены пароля
	CreatedAt   string `json:"created_at"`
}
