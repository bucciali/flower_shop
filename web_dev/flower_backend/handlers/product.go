package handlers

import (
	"encoding/json"
	"flower_backend/db"
	"flower_backend/models"
	"net/http"
	"strconv"
	"strings"
)

// GET /api/products
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	rows, err := db.DB.Query("SELECT id, name, price, description, image_url, category, is_available, created_at FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.ImageURL, &p.Category, &p.IsAvailable, &p.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	json.NewEncoder(w).Encode(products)
}

// POST /api/products
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	query := `
		INSERT INTO products (name, price, description, image_url, category, is_available, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	if err := db.DB.QueryRow(query, p.Name, p.Price, p.Description, p.ImageURL, p.Category, p.IsAvailable).
		Scan(&p.ID, &p.CreatedAt); err != nil {
		http.Error(w, "Ошибка при сохранении продукта: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// GET /api/products/{id}
func GetProductByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "ID продукта не указан", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Неверный ID продукта", http.StatusBadRequest)
		return
	}

	var p models.Product
	err = db.DB.QueryRow(
		"SELECT id, name, price, description, image_url, category, is_available, created_at FROM products WHERE id=$1", id).
		Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.ImageURL, &p.Category, &p.IsAvailable, &p.CreatedAt)
	if err != nil {
		http.Error(w, "Продукт не найден", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(p)
}

// PUT /api/products/{id}
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "ID продукта не указан", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Неверный ID продукта", http.StatusBadRequest)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	query := `
		UPDATE products 
		SET name=$1, price=$2, description=$3, image_url=$4, category=$5, is_available=$6
		WHERE id=$7
		RETURNING created_at`
	if err := db.DB.QueryRow(query, p.Name, p.Price, p.Description, p.ImageURL, p.Category, p.IsAvailable, id).
		Scan(&p.CreatedAt); err != nil {
		http.Error(w, "Ошибка обновления продукта: "+err.Error(), http.StatusInternalServerError)
		return
	}

	p.ID = id
	json.NewEncoder(w).Encode(p)
}

// DELETE /api/products/{id}
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "ID продукта не указан", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Неверный ID продукта", http.StatusBadRequest)
		return
	}

	res, err := db.DB.Exec("DELETE FROM products WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Ошибка удаления продукта: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Продукт не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
