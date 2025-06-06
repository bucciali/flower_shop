package routes

import (
	"flower_backend/handlers"
	"net/http"
	"strings"
)

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/api/products" {
			handlers.GetAllProducts(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/products/") {
			handlers.GetProductByID(w, r)
		} else {
			http.NotFound(w, r)
		}
	} else if r.Method == http.MethodPost {
		if r.URL.Path == "/api/products" {
			handlers.CreateProduct(w, r)
		} else {
			http.NotFound(w, r)
		}
	} else if r.Method == http.MethodPut {
		if strings.HasPrefix(r.URL.Path, "/api/products/") {
			handlers.UpdateProduct(w, r)
		} else {
			http.NotFound(w, r)
		}
	} else if r.Method == http.MethodDelete {
		if strings.HasPrefix(r.URL.Path, "/api/products/") {
			handlers.DeleteProduct(w, r)
		} else {
			http.NotFound(w, r)
		}
	} else {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
