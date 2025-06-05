package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Coderovshik/meet/internal/auth"
)

func HandleRegister(us *auth.UserStore, ls *auth.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if creds.Username == "" || creds.Password == "" {
			http.Error(w, "Username and password are required", http.StatusBadRequest)
			return
		}
		if err := us.CreateUser(r.Context(), creds.Username, creds.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Логируем успешную регистрацию
		details := fmt.Sprintf("IP: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())
		if err := ls.AddLog(r.Context(), creds.Username, "registration", details); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			fmt.Printf("Ошибка при логировании регистрации: %v\n", err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func HandleLogin(us *auth.UserStore, ls *auth.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		valid, err := us.ValidateUser(r.Context(), creds.Username, creds.Password)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if !valid {
			// Логируем неудачную попытку входа
			details := fmt.Sprintf("Неудачная попытка входа. IP: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())
			if err := ls.AddLog(r.Context(), creds.Username, "login_failed", details); err != nil {
				fmt.Printf("Ошибка при логировании неудачной попытки входа: %v\n", err)
			}

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Логируем успешный вход
		details := fmt.Sprintf("IP: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())
		if err := ls.AddLog(r.Context(), creds.Username, "login", details); err != nil {
			fmt.Printf("Ошибка при логировании входа: %v\n", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

// HandleGetUserLogs обрабатывает запрос на получение логов пользователя
func HandleGetUserLogs(ls *auth.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем имя пользователя из контекста запроса
		username, ok := auth.GetUsernameFromContext(r.Context())
		if !ok {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		// Получаем лимит из параметров запроса, по умолчанию 50
		limitStr := r.URL.Query().Get("limit")
		limit := int64(50)
		if limitStr != "" {
			var err error
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil || limit <= 0 {
				http.Error(w, "Неверный параметр limit", http.StatusBadRequest)
				return
			}
		}

		// Получаем логи пользователя
		logs, err := ls.GetLogs(r.Context(), username, limit)
		if err != nil {
			http.Error(w, "Ошибка при получении логов", http.StatusInternalServerError)
			return
		}

		// Отправляем ответ клиенту
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(logs); err != nil {
			http.Error(w, "Ошибка при сериализации ответа", http.StatusInternalServerError)
			return
		}
	}
}
