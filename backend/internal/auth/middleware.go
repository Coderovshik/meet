package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// UsernameContextKey используется для хранения имени пользователя в контексте запроса
	UsernameContextKey contextKey = "username"
)

// GetUsernameFromContext извлекает имя пользователя из контекста запроса
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameContextKey).(string)
	return username, ok
}

// AuthMiddleware создает middleware для аутентификации пользователя
func AuthMiddleware(us *UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен авторизации из заголовка
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Не предоставлен токен авторизации", http.StatusUnauthorized)
				return
			}

			// Извлекаем имя пользователя и пароль из заголовка
			// Формат: "Basic username:password"
			if !strings.HasPrefix(authHeader, "Basic ") {
				http.Error(w, "Неправильный формат токена авторизации", http.StatusUnauthorized)
				return
			}

			// Разбиваем строку на части и получаем пару username:password
			parts := strings.SplitN(authHeader[6:], ":", 2)
			if len(parts) != 2 {
				http.Error(w, "Неправильный формат токена авторизации", http.StatusUnauthorized)
				return
			}

			username, password := parts[0], parts[1]

			// Проверяем валидность учетных данных
			valid, err := us.ValidateUser(r.Context(), username, password)
			if err != nil {
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
			if !valid {
				http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
				return
			}

			// Добавляем имя пользователя в контекст запроса
			ctx := context.WithValue(r.Context(), UsernameContextKey, username)

			// Вызываем следующий обработчик с обновленным контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
