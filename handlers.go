package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

// RegisterHandler обрабатывает регистрацию нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := parseJSONRequest(r, &req); err != nil {
		sendErrorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := validateRegisterRequest(&req); err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := UserExistsByEmail(req.Email)
	if err != nil {
		log.Printf("check user exists: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		sendErrorResponse(w, "user already exists", http.StatusConflict)
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("hash password: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := CreateUser(req.Email, req.Username, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			sendErrorResponse(w, "user already exists", http.StatusConflict)
			return
		}
		log.Printf("create user: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(*user)
	if err != nil {
		log.Printf("generate token: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	sendJSONResponse(w, AuthResponse{Token: token, User: *user}, http.StatusCreated)
}

// LoginHandler обрабатывает вход пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := parseJSONRequest(r, &req); err != nil {
		sendErrorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := validateLoginRequest(&req); err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		log.Printf("get user by email: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !CheckPassword(req.Password, user.PasswordHash) {
		sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(*user)
	if err != nil {
		log.Printf("generate token: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	sendJSONResponse(w, AuthResponse{Token: token, User: *user}, http.StatusOK)
}

// ProfileHandler возвращает профиль текущего пользователя
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		sendErrorResponse(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorResponse(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("get user by id: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, user, http.StatusOK)
}

// HealthHandler проверяет состояние сервиса
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем подключение к БД
	if db != nil {
		if err := db.Ping(); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}
	}

	// Возвращаем статус OK
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "ok",
		"message": "Service is running",
	}
	json.NewEncoder(w).Encode(response)
}

// sendJSONResponse отправляет JSON ответ (вспомогательная функция)
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse отправляет JSON ответ с ошибкой (вспомогательная функция)
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

// parseJSONRequest парсит JSON из тела запроса (вспомогательная функция)
func parseJSONRequest(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Строгая проверка полей

	return decoder.Decode(v)
}

// validateRegisterRequest валидирует данные регистрации
func validateRegisterRequest(req *RegisterRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if len(req.Username) < 3 || len(req.Username) > 30 {
		return fmt.Errorf("username must be between 3 and 30 characters")
	}
	for _, r := range req.Username {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return fmt.Errorf("username may contain only letters, digits and underscore")
		}
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return true
	}
	return strings.Contains(err.Error(), "duplicate key")
}

// validateLoginRequest валидирует данные входа
func validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
