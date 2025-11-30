package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Liana-wq1/my-first-go/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// тело запроса на регистрацию
type RegisterRequest struct {
	FirstName string `json:"name"`    //в register.html мы отправляли name, поэтому тут FirstName маппим на name.
	LastName  string `json:"surname"` //в register.html мы отправляли surname, поэтому тут LastName маппим на surname.
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// тело запроса на логин
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request, sender *EmailSender) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST request", http.StatusMethodNotAllowed) //неправильный метод
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Warning JSON", http.StatusBadRequest) //некорректный JSON
		return
	}

	if req.FirstName == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest) //не все обязательные поля заполнены
		return
	}

	// хэшируем пароль
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка шифрования пароля", http.StatusInternalServerError)
		return
	}
	hash := string(hashBytes)

	// генерируем токен подтверждения
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	now := time.Now()

	// создаём пользователя в таблице users
	res, err := storage.DB.Exec(`
		INSERT INTO users(first_name, last_name, email, phone, avatar_url, created_at, updated_at, is_confirmed, confirm_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)
	`, req.FirstName, "", req.Email, "", "", now, now, token)
	if err != nil {
		http.Error(w, "Такой пользователь уже существует", http.StatusBadRequest)
		return
	}

	userID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// создаём запись в credentials
	_, err = storage.DB.Exec(`
		INSERT INTO credentials(user_id, password_hash, last_login_at)
		VALUES (?, ?, NULL)
	`, userID, hash)
	if err != nil {
		http.Error(w, "Ошибка сохранения пароля", http.StatusInternalServerError)
		return
	}

	// Отправка письма
	// формируем ссылку подтверждения
	link := "http://localhost:8080/api/confirm?token=" + token

	// тело письма
	subject := "Подтверждение регистрации"
	body := BuildConfirmEmailBody(link)

	// отправляем письмо
	if err := sender.Send(req.Email, subject, body); err != nil {
		http.Error(w, "Регистрация прошла, но письмо не удалось отправить: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Регистрация успешна! На вашу почту направлено письмо, перейдите по ссылке для подтверждения email"))
}

func ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Токен не передан", http.StatusBadRequest)
		return
	}

	res, err := storage.DB.Exec(`
		UPDATE users
		SET is_confirmed = 1, confirm_token = NULL
		WHERE confirm_token = ?
	`, token)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		http.Error(w, "Неверный или устаревший токен", http.StatusBadRequest)
		return
	}

	w.Write([]byte("Email подтверждён! Теперь вы можете войти."))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	// читаем хэш пароля и статус подтверждения
	row := storage.DB.QueryRow(`
		SELECT u.id, u.is_confirmed, c.password_hash
		FROM users u
		JOIN credentials c ON u.id = c.user_id
		WHERE u.email = ?
	`, req.Email)

	var (
		userID       int64
		isConfirmed  int
		passwordHash string
	)

	err := row.Scan(&userID, &isConfirmed, &passwordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "Пользователь не найден", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	if isConfirmed == 0 {
		http.Error(w, "Email не подтверждён", http.StatusForbidden)
		return
	}

	// сравнение пароля
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// логически тут нужно выдать сессию/JWT, но пока просто текст:
	w.Write([]byte("Успешный вход! (user_id=" + fmt.Sprint(userID) + ")"))
}
