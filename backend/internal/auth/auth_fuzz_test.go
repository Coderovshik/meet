package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// Настройка тестового окружения с miniredis
func setupTestEnv(t *testing.T) (*UserStore, *LogStore, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Ошибка при запуске miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	userStore := NewUserStore(client)
	logStore := NewLogStore(client)

	return userStore, logStore, mr
}

// FuzzCreateUser тестирует функцию создания пользователя с различными входными данными
func FuzzCreateUser(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("user1", "pass123")
	f.Add("usr", "pass123") // Короткое имя
	f.Add("user1", "pas")   // Короткий пароль
	f.Add("", "pass123")
	f.Add("user1", "")
	f.Add("123", "pass123")
	f.Add("user1", "123")
	f.Add("user!@#", "pass123") // Недопустимые символы
	f.Add("user1", "pass!@#")

	f.Fuzz(func(t *testing.T, username, password string) {
		userStore, _, mr := setupTestEnv(t)
		defer mr.Close()

		// Пытаемся создать пользователя, проверяем только что функция не паникует
		_ = userStore.CreateUser(context.Background(), username, password)
	})
}

// FuzzValidateUser тестирует функцию валидации пользователя с различными входными данными
func FuzzValidateUser(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("user1", "pass123")
	f.Add("", "pass123")
	f.Add("user1", "")
	f.Add("usr", "pass123")     // Короткое имя
	f.Add("user1", "pas")       // Короткий пароль
	f.Add("user!@#", "pass123") // Недопустимые символы
	f.Add("user1", "pass!@#")

	f.Fuzz(func(t *testing.T, username, password string) {
		userStore, _, mr := setupTestEnv(t)
		defer mr.Close()

		// Создаем тестового пользователя
		if len(username) >= 4 && len(password) >= 4 {
			err := userStore.CreateUser(context.Background(), "validuser", "validpass")
			if err != nil {
				t.Skip("Не удалось создать тестового пользователя")
			}
		}

		// Проверяем валидность пользователя, просто убеждаемся что функция не паникует
		_, _ = userStore.ValidateUser(context.Background(), username, password)
	})
}

// FuzzAuthMiddleware тестирует middleware авторизации с различными заголовками
func FuzzAuthMiddleware(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("Basic user1:pass123")
	f.Add("Basic :")
	f.Add("Basic user1:")
	f.Add("Basic :pass123")
	f.Add("Bearer token123")
	f.Add("")
	f.Add("user1:pass123")
	f.Add("Basic user1:pass123:extra")
	f.Add("Basic " + string([]byte{0xff, 0xfe})) // Невалидные UTF-8 байты

	f.Fuzz(func(t *testing.T, authHeader string) {
		userStore, _, mr := setupTestEnv(t)
		defer mr.Close()

		// Создаем тестового пользователя
		err := userStore.CreateUser(context.Background(), "user1", "pass123")
		if err != nil {
			t.Skip("Не удалось создать тестового пользователя")
		}

		// Создаем тестовый запрос
		req, err := http.NewRequest("GET", "/api/logs", nil)
		if err != nil {
			t.Skip("Ошибка создания запроса")
		}
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Создаем тестовый обработчик, который будет обернут middleware
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, ok := GetUsernameFromContext(r.Context())
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(username))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		})

		// Оборачиваем тестовый обработчик в middleware
		handler := AuthMiddleware(userStore)(testHandler)

		// Выполняем запрос, проверяем что middleware не паникует
		handler.ServeHTTP(rr, req)
	})
}

// FuzzAddLog тестирует функцию добавления лога с различными входными данными
func FuzzAddLog(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("user1", "login", "успешный вход")
	f.Add("", "login", "успешный вход")
	f.Add("user1", "", "успешный вход")
	f.Add("user1", "login", "")
	f.Add("user!@#", "login", "успешный вход") // Специальные символы
	f.Add("user1", "login!@#", "успешный вход")
	f.Add("user1", "login", "успешный вход!@#")
	f.Add(string([]byte{0xff, 0xfe}), "login", "успешный вход") // Невалидные UTF-8 байты
	f.Add("user1", string([]byte{0xff, 0xfe}), "успешный вход")
	f.Add("user1", "login", string([]byte{0xff, 0xfe}))

	f.Fuzz(func(t *testing.T, username, action, details string) {
		_, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Пытаемся добавить лог, проверяем только что функция не паникует
		_ = logStore.AddLog(context.Background(), username, action, details)
	})
}

// FuzzGetLogs тестирует функцию получения логов с различными входными данными
func FuzzGetLogs(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("user1", int64(10))
	f.Add("", int64(10))
	f.Add("user!@#", int64(10)) // Специальные символы
	f.Add("user1", int64(0))
	f.Add("user1", int64(-1))
	f.Add("user1", int64(9999999))
	f.Add(string([]byte{0xff, 0xfe}), int64(10)) // Невалидные UTF-8 байты

	f.Fuzz(func(t *testing.T, username string, limit int64) {
		_, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Добавляем тестовый лог
		if username != "" {
			err := logStore.AddLog(context.Background(), "testuser", "test", "test log")
			if err != nil {
				t.Skip("Не удалось добавить тестовый лог")
			}
		}

		// Пытаемся получить логи, проверяем только что функция не паникует
		_, _ = logStore.GetLogs(context.Background(), username, limit)
	})
}

// FuzzGetLogsByTimeRange тестирует функцию получения логов по временному диапазону
func FuzzGetLogsByTimeRange(f *testing.F) {
	// Добавляем начальные корпусы для фаззера с различными временными диапазонами в формате строк
	f.Add("user1", "2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z")
	f.Add("", "2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z")
	f.Add("user1", "2023-12-31T23:59:59Z", "2023-01-01T00:00:00Z")                    // Некорректный диапазон (конец раньше начала)
	f.Add("user1", "", "2023-12-31T23:59:59Z")                                        // Пустое время начала
	f.Add("user1", "2023-01-01T00:00:00Z", "")                                        // Пустое время конца
	f.Add("user1", "invalid-date", "2023-12-31T23:59:59Z")                            // Невалидная дата начала
	f.Add("user1", "2023-01-01T00:00:00Z", "invalid-date")                            // Невалидная дата конца
	f.Add("user!@#", "2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z")                  // Специальные символы в имени
	f.Add(string([]byte{0xff, 0xfe}), "2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z") // Невалидные UTF-8 байты в имени

	f.Fuzz(func(t *testing.T, username, startStr, endStr string) {
		_, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Конвертируем строковые даты в time.Time
		var start, end time.Time
		var err error

		if startStr != "" {
			start, err = time.Parse(time.RFC3339, startStr)
			if err != nil {
				// Если парсинг не удался, используем нулевое время
				start = time.Time{}
			}
		}

		if endStr != "" {
			end, err = time.Parse(time.RFC3339, endStr)
			if err != nil {
				// Если парсинг не удался, используем текущее время
				end = time.Now()
			}
		}

		// Добавляем тестовый лог
		if username != "" {
			err := logStore.AddLog(context.Background(), "testuser", "test", "test log")
			if err != nil {
				t.Skip("Не удалось добавить тестовый лог")
			}
		}

		// Пытаемся получить логи по временному диапазону, проверяем только что функция не паникует
		_, _ = logStore.GetLogsByTimeRange(context.Background(), username, start, end)
	})
}

// FuzzClearLogs тестирует функцию очистки логов
func FuzzClearLogs(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("user1")
	f.Add("")
	f.Add("user!@#")                  // Специальные символы
	f.Add(string([]byte{0xff, 0xfe})) // Невалидные UTF-8 байты

	f.Fuzz(func(t *testing.T, username string) {
		_, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Добавляем тестовый лог
		if username != "" {
			err := logStore.AddLog(context.Background(), username, "test", "test log")
			if err != nil {
				t.Skip("Не удалось добавить тестовый лог")
			}
		}

		// Пытаемся очистить логи, проверяем только что функция не паникует
		_ = logStore.ClearLogs(context.Background(), username)
	})
}

// FuzzGetUsernameFromContext тестирует функцию получения имени пользователя из контекста
func FuzzGetUsernameFromContext(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add("user1")
	f.Add("")
	f.Add("user!@#")                  // Специальные символы
	f.Add(string([]byte{0xff, 0xfe})) // Невалидные UTF-8 байты

	f.Fuzz(func(t *testing.T, username string) {
		// Создаем контекст с различными значениями username
		ctx := context.WithValue(context.Background(), UsernameContextKey, username)

		// Пытаемся получить имя пользователя из контекста
		extractedUsername, ok := GetUsernameFromContext(ctx)
		if ok && extractedUsername != username {
			t.Errorf("Извлеченное имя пользователя %q не соответствует установленному %q",
				extractedUsername, username)
		}

		// Проверяем также с контекстом без имени пользователя
		_, ok = GetUsernameFromContext(context.Background())
		if ok {
			t.Error("GetUsernameFromContext вернул ok=true для контекста без имени пользователя")
		}
	})
}
