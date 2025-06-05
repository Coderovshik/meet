package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
}

type LogStore struct {
	client *redis.Client
}

func NewLogStore(client *redis.Client) *LogStore {
	return &LogStore{client: client}
}

// AddLog добавляет новую запись в лог пользователя
func (ls *LogStore) AddLog(ctx context.Context, username string, action, details string) error {
	key := "logs:" + username

	entry := LogEntry{
		Timestamp: time.Now(),
		Action:    action,
		Details:   details,
	}

	// Сериализуем запись в JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Добавляем запись в список логов пользователя
	// Используем RPUSH для добавления в конец списка
	return ls.client.RPush(ctx, key, entryJSON).Err()
}

// GetLogs получает последние N записей из лога пользователя
func (ls *LogStore) GetLogs(ctx context.Context, username string, limit int64) ([]LogEntry, error) {
	key := "logs:" + username

	// Получаем последние N записей
	// Используем LRANGE для получения элементов из списка
	entries, err := ls.client.LRange(ctx, key, -limit, -1).Result()
	if err != nil {
		return nil, err
	}

	logs := make([]LogEntry, 0, len(entries))
	for _, entry := range entries {
		var logEntry LogEntry
		if err := json.Unmarshal([]byte(entry), &logEntry); err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}

// ClearLogs очищает все логи пользователя
func (ls *LogStore) ClearLogs(ctx context.Context, username string) error {
	key := "logs:" + username
	return ls.client.Del(ctx, key).Err()
}

// GetLogsByTimeRange получает логи пользователя за определенный период времени
func (ls *LogStore) GetLogsByTimeRange(ctx context.Context, username string, start, end time.Time) ([]LogEntry, error) {
	key := "logs:" + username

	// Получаем все записи
	entries, err := ls.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	logs := make([]LogEntry, 0)
	for _, entry := range entries {
		var logEntry LogEntry
		if err := json.Unmarshal([]byte(entry), &logEntry); err != nil {
			return nil, err
		}

		// Фильтруем по временному диапазону
		if logEntry.Timestamp.After(start) && logEntry.Timestamp.Before(end) {
			logs = append(logs, logEntry)
		}
	}

	return logs, nil
}
