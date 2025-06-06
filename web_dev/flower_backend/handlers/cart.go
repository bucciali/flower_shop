package handlers

import (
	"encoding/json"
	"errors"
	"flower_backend/db"
	"net/http"
	"strconv"
)

type CartItemInput struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func parseCartItem(r *http.Request) (CartItemInput, error) {
	var item CartItemInput
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		return item, errors.New("неверный формат JSON")
	}
	if item.UserID <= 0 || item.ProductID <= 0 {
		return item, errors.New("user_id и product_id должны быть положительными")
	}
	return item, nil
}

// POST /api/cart/add
func AddToCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	item, err := parseCartItem(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if item.Quantity <= 0 {
		http.Error(w, "Количество должно быть больше 0", http.StatusBadRequest)
		return
	}

	var exists bool
	err = db.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM cart WHERE user_id=$1 AND product_id=$2)",
		item.UserID, item.ProductID,
	).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка проверки корзины", http.StatusInternalServerError)
		return
	}

	if exists {
		_, err = db.DB.Exec(
			"UPDATE cart SET quantity = quantity + $1 WHERE user_id=$2 AND product_id=$3",
			item.Quantity, item.UserID, item.ProductID,
		)
	} else {
		_, err = db.DB.Exec(
			"INSERT INTO cart (user_id, product_id, quantity) VALUES ($1, $2, $3)",
			item.UserID, item.ProductID, item.Quantity,
		)
	}
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Товар добавлен в корзину",
	})
}

// GET /api/cart?user_id={id}
func GetCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Нужен параметр user_id", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		http.Error(w, "user_id должен быть числом", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query(`
		SELECT p.id, p.name, p.price, c.quantity, p.image_url 
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1`, userID)
	if err != nil {
		http.Error(w, "Ошибка запроса корзины", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, quantity int
		var name, imageURL string
		var price float64
		if err := rows.Scan(&id, &name, &price, &quantity, &imageURL); err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		items = append(items, map[string]interface{}{
			"product_id": id,
			"name":       name,
			"price":      price,
			"quantity":   quantity,
			"image":      imageURL,
		})
	}

	json.NewEncoder(w).Encode(items)
}

// DELETE /api/cart/remove
func RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var item struct {
		UserID    int `json:"user_id"`
		ProductID int `json:"product_id"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("DELETE FROM cart WHERE user_id=$1 AND product_id=$2", item.UserID, item.ProductID)
	if err != nil {
		http.Error(w, "Ошибка удаления из корзины", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Товар удалён из корзины",
	})
}

// DELETE /api/cart/clear?user_id={id}
func ClearCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Нужен user_id", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("DELETE FROM cart WHERE user_id = $1", userIDStr)
	if err != nil {
		http.Error(w, "Ошибка очистки корзины", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Корзина очищена",
	})
}

// PUT /api/cart/update
func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPut {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	item, err := parseCartItem(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if item.Quantity <= 0 {
		http.Error(w, "Количество должно быть больше 0", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(
		"UPDATE cart SET quantity=$1 WHERE user_id=$2 AND product_id=$3",
		item.Quantity, item.UserID, item.ProductID,
	)
	if err != nil {
		http.Error(w, "Ошибка обновления корзины", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Количество обновлено",
	})
}
