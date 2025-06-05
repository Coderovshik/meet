package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Coderovshik/meet/internal/api"
	"github.com/Coderovshik/meet/internal/auth"
	"github.com/Coderovshik/meet/internal/signaling"

	"github.com/redis/go-redis/v9"
)

func main() {
	redis_host := os.Getenv("REDIS_HOST")
	if redis_host == "" {
		redis_host = "localhost"
	}
	redisClient := redis.NewClient(&redis.Options{
		// Addr: fmt.Sprintf("%s:41163", redis_host),
		Addr: fmt.Sprintf("%s:6379", redis_host),
	})
	userStore := auth.NewUserStore(redisClient)
	logStore := auth.NewLogStore(redisClient)

	http.HandleFunc("/api/register", api.HandleRegister(userStore, logStore))
	http.HandleFunc("/api/login", api.HandleLogin(userStore, logStore))
	http.Handle("/ws", signaling.HandleWebSocket(userStore, logStore))

	logsHandler := http.HandlerFunc(api.HandleGetUserLogs(logStore))

	http.Handle("/api/logs", auth.AuthMiddleware(userStore)(logsHandler))

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// log.Println("Server running on :80")
	// log.Fatal(http.ListenAndServe(":80", nil))
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
