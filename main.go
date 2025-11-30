package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Liana-wq1/my-first-go/internal/service"
	"github.com/Liana-wq1/my-first-go/internal/storage"
	"github.com/joho/godotenv"
)

func main() {

	// Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: .env file not found, using system env")
	}

	storage.InitSQLite()

	emailSender := service.NewEmailSender(
		os.Getenv("SMTP_EMAIL"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
	)

	// статика (фронт)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// API
	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		service.RegisterHandler(w, r, emailSender)
	})
	http.HandleFunc("/api/login", service.LoginHandler)
	http.HandleFunc("/api/confirm", service.ConfirmHandler)

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
