package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Coderovshik/meet/internal/auth"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// Настройка минимального тестового окружения с миниредисом
func setupTestEnv(t *testing.T) (*auth.UserStore, *auth.LogStore, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Ошибка при запуске miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	userStore := auth.NewUserStore(client)
	logStore := auth.NewLogStore(client)

	return userStore, logStore, mr
}

// FuzzRegisterHandler проверяет обработчик регистрации с разными входными данными
func FuzzRegisterHandler(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add([]byte(`{"username":"user1","password":"pass123"}`))
	f.Add([]byte(`{"username":"","password":"pass123"}`))
	f.Add([]byte(`{"username":"user1","password":""}`))
	f.Add([]byte(`{"username":123,"password":"pass123"}`))
	f.Add([]byte(`{"username":"user1","password":123}`))
	f.Add([]byte(`{"username":"usr","password":"pass123"}`)) // Слишком короткое имя
	f.Add([]byte(`{"username":"user1","password":"pas"}`))   // Слишком короткий пароль
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`"string"`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		userStore, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Создаем тестовый запрос
		req, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer(data))
		if err != nil {
			return // Пропускаем невалидные запросы
		}
		req.Header.Set("Content-Type", "application/json")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Создаем обработчик
		handler := HandleRegister(userStore, logStore)

		// Выполняем запрос
		handler.ServeHTTP(rr, req)

		// Просто проверяем, что обработчик не паникует
		// Конкретные проверки кодов ответа здесь не делаем, так как это фаззинг-тест
	})
}

// FuzzLoginHandler проверяет обработчик логина с разными входными данными
func FuzzLoginHandler(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add([]byte(`{"username":"user1","password":"pass123"}`))
	f.Add([]byte(`{"username":"","password":"pass123"}`))
	f.Add([]byte(`{"username":"user1","password":""}`))
	f.Add([]byte(`{"username":123,"password":"pass123"}`))
	f.Add([]byte(`{"username":"user1","password":123}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`"string"`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		userStore, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Создаем тестовый запрос
		req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer(data))
		if err != nil {
			return // Пропускаем невалидные запросы
		}
		req.Header.Set("Content-Type", "application/json")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Создаем обработчик
		handler := HandleLogin(userStore, logStore)

		// Выполняем запрос
		handler.ServeHTTP(rr, req)

		// Просто проверяем, что обработчик не паникует
	})
}

// FuzzAuthMiddleware проверяет middleware авторизации с различными заголовками
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

		// Создаем тестовый запрос
		req, err := http.NewRequest("GET", "/api/logs", nil)
		if err != nil {
			return // Пропускаем невалидные запросы
		}
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Создаем тестовый обработчик, который будет обернут middleware
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Оборачиваем тестовый обработчик в middleware
		handler := auth.AuthMiddleware(userStore)(testHandler)

		// Выполняем запрос
		handler.ServeHTTP(rr, req)

		// Просто проверяем, что middleware не паникует
	})
}

// FuzzGetUserLogs проверяет обработчик получения логов с различными параметрами запроса
func FuzzGetUserLogs(f *testing.F) {
	// Добавляем начальные корпусы для фаззера с различными параметрами limit
	f.Add("50")
	f.Add("0")
	f.Add("-1")
	f.Add("9999999999999")
	f.Add("abc")
	f.Add("")
	f.Add("3.14")
	f.Add(string([]byte{0xff, 0xfe})) // Невалидные UTF-8 байты

	f.Fuzz(func(t *testing.T, limitParam string) {
		_, logStore, mr := setupTestEnv(t)
		defer mr.Close()

		// Создаем тестовый запрос
		url := "/api/logs"
		if limitParam != "" {
			url += "?limit=" + limitParam
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return // Пропускаем невалидные запросы
		}

		// Устанавливаем контекст с имитацией аутентифицированного пользователя
		ctx := context.WithValue(req.Context(), auth.UsernameContextKey, "testuser")
		req = req.WithContext(ctx)

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Создаем обработчик
		handler := HandleGetUserLogs(logStore)

		// Выполняем запрос
		handler.ServeHTTP(rr, req)

		// Просто проверяем, что обработчик не паникует
	})
}

// Тестируем с различными структурами LogEntry для сериализации/десериализации
func FuzzLogEntryJSON(f *testing.F) {
	// Добавляем начальные корпусы для фаззера
	f.Add([]byte(`{"timestamp":"2023-05-01T12:00:00Z","action":"login","details":"successful login"}`))
	f.Add([]byte(`{"timestamp":"invalid-time","action":"login","details":"successful login"}`))
	f.Add([]byte(`{"timestamp":"2023-05-01T12:00:00Z","action":123,"details":"successful login"}`))
	f.Add([]byte(`{"timestamp":"2023-05-01T12:00:00Z","action":"login","details":123}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		var entry auth.LogEntry

		// Пытаемся десериализовать JSON в структуру LogEntry
		// Нас интересует только то, что процесс не вызывает панику
		_ = json.Unmarshal(data, &entry)

		// Также проверяем обратную сериализацию
		_, _ = json.Marshal(entry)
	})
}
