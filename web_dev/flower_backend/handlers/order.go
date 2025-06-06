package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"flower_backend/db"

	"github.com/lib/pq"
)

// PlaceOrder — POST /api/cart/checkout
func PlaceOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// на входе JSON {"user_id": 11}
	var input struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 1) Начинаем транзакцию
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Ошибка начала транзакции", http.StatusInternalServerError)
		return
	}

	// 2) Собираем все товары из cart + считаем общую сумму
	rows, err := tx.Query(`
		SELECT c.product_id, c.quantity, p.price
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1
	`, input.UserID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка чтения корзины", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type CartRow struct {
		ProductID int
		Quantity  int
		Price     float64
	}

	var (
		items      []CartRow
		totalPrice float64
	)

	for rows.Next() {
		var it CartRow
		if err := rows.Scan(&it.ProductID, &it.Quantity, &it.Price); err != nil {
			tx.Rollback()
			http.Error(w, "Ошибка чтения товара из корзины", http.StatusInternalServerError)
			return
		}
		items = append(items, it)
		totalPrice += float64(it.Quantity) * it.Price
	}

	// Если в корзине нет товаров — откатываем
	if len(items) == 0 {
		tx.Rollback()
		http.Error(w, "Корзина пуста", http.StatusBadRequest)
		return
	}

	// 3) Собираем массив product_ids (без учёта количества, просто ID)
	var productIDs []int
	for _, it := range items {
		productIDs = append(productIDs, it.ProductID)
	}

	// 4) Вставляем в таблицу orders (product_ids INT[])
	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (user_id, product_ids, total_price, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`,
		input.UserID,
		pq.Array(productIDs), // pq.Array конвертирует Go-[]int в SQL INT[]
		totalPrice,
		"pending",
		time.Now(),
	).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка создания заказа: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5) Очищаем корзину пользователя
	if _, err = tx.Exec("DELETE FROM cart WHERE user_id = $1", input.UserID); err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка очистки корзины", http.StatusInternalServerError)
		return
	}

	// 6) Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		http.Error(w, "Ошибка сохранения заказа", http.StatusInternalServerError)
		return
	}

	// 7) Возвращаем финальный JSON
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Заказ оформлен",
		"order_id":    orderID,
		"total":       totalPrice,
		"product_ids": productIDs,
	})
}
