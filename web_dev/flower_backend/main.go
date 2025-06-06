// flower_backend/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"flower_backend/db"
	"flower_backend/handlers"
	"flower_backend/routes"

	"github.com/joho/godotenv"
)

func main() {
	// 1) Загрузить .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// 2) Установить соединение с БД
	db.Init()

	// 3) Зарегистрировать API-роуты
	http.HandleFunc("/api/products", routes.ProductsHandler)
	http.HandleFunc("/api/products/", routes.ProductsHandler)

	http.HandleFunc("/api/register", handlers.RegisterHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)

	http.HandleFunc("/api/user/profile", handlers.GetProfile)
	http.HandleFunc("/api/user/update", handlers.UpdateProfile)
	http.HandleFunc("/api/user/password", handlers.ChangePassword)

	http.HandleFunc("/api/cart/add", handlers.AddToCart)
	http.HandleFunc("/api/cart", handlers.GetCart)
	http.HandleFunc("/api/cart/remove", handlers.RemoveFromCart)
	http.HandleFunc("/api/cart/clear", handlers.ClearCart)
	http.HandleFunc("/api/cart/update", handlers.UpdateCartItem)
	http.HandleFunc("/api/cart/checkout", handlers.PlaceOrder)

	// 4) Всё, что не /api/…, отдаём из папки my_site
	fileServer := http.FileServer(http.Dir("../my_site"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Если запрос начинается с /api/, то вернём 404 по умолчанию (его уже обрабатывают выше).
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Иначе – просто отдать файл из ../my_site/<whatever>
		fileServer.ServeHTTP(w, r)
	})

	// 5) Запустить HTTP
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
