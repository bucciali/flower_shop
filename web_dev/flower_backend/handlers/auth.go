package handlers

import (
	"database/sql"
	"encoding/json"
	"flower_backend/db"
	"flower_backend/models"
	"net/http"
	"strconv"
)

// POST /api/register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Пользователь с таким email уже существует", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO users (username, email, password, created_at) 
	          VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`
	var newID int
	var createdAt string
	err = db.DB.QueryRow(query, req.Username, req.Email, req.Password).Scan(&newID, &createdAt)
	if err != nil {
		http.Error(w, "Ошибка при сохранении пользователя", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"user_id":    newID,
		"username":   req.Username,
		"email":      req.Email, // возвращаем email сразу
		"created_at": createdAt,
		"message":    "Регистрация прошла успешно",
	}
	json.NewEncoder(w).Encode(resp)
}

// POST /api/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var user models.User
	var passwordFromDB string
	err := db.DB.QueryRow(
		"SELECT id, username, password, created_at FROM users WHERE email=$1",
		req.Email,
	).Scan(&user.ID, &user.Username, &passwordFromDB, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Неверные email или пароль", http.StatusUnauthorized)
		} else {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		}
		return
	}

	if req.Password != passwordFromDB {
		http.Error(w, "Неверные email или пароль", http.StatusUnauthorized)
		return
	}

	user.Email = req.Email // заполняем поле email для ответа
	user.Password = ""     // не возвращаем пароль

	resp := map[string]interface{}{
		"user_id":    user.ID,
		"username":   user.Username,
		"email":      user.Email, // теперь включаем email
		"created_at": user.CreatedAt,
		"message":    "Вход выполнен успешно",
	}
	json.NewEncoder(w).Encode(resp)
}

// GET /api/user/profile?user_id={id}
func GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Нужен параметр user_id", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		http.Error(w, "Некорректный user_id", http.StatusBadRequest)
		return
	}

	var user models.User
	err = db.DB.QueryRow(
		"SELECT id, username, email, created_at FROM users WHERE id=$1",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(user)
}

// PUT /api/user/update
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPut {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.UserID <= 0 || req.Username == "" || req.Email == "" {
		http.Error(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	res, err := db.DB.Exec(
		"UPDATE users SET username=$1, email=$2 WHERE id=$3",
		req.Username, req.Email, req.UserID,
	)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	resp := map[string]string{"message": "Профиль обновлён"}
	json.NewEncoder(w).Encode(resp)
}

// PUT /api/user/password
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPut {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID      int    `json:"user_id"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.UserID <= 0 || req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	var currentPassword string
	err := db.DB.QueryRow(
		"SELECT password FROM users WHERE id=$1",
		req.UserID,
	).Scan(&currentPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		}
		return
	}

	if currentPassword != req.OldPassword {
		http.Error(w, "Старый пароль неверный", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET password=$1 WHERE id=$2", req.NewPassword, req.UserID)
	if err != nil {
		http.Error(w, "Ошибка обновления пароля", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"message": "Пароль успешно изменён"}
	json.NewEncoder(w).Encode(resp)
}
