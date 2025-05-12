package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Coderovshik/meet/internal/api"
	"github.com/Coderovshik/meet/internal/auth"
	"github.com/Coderovshik/meet/internal/rooms"
	"github.com/Coderovshik/meet/internal/signaling"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Инициализация Redis
	redis_host := os.Getenv("REDIS_HOST")
	if redis_host == "" {
		redis_host = "localhost"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:6379", redis_host),
	})
	userStore := auth.NewUserStore(redisClient)

	roomManager := rooms.NewManager()

	// Маршруты
	http.HandleFunc("/api/register", api.HandleRegister(userStore))
	http.HandleFunc("/api/login", api.HandleLogin(userStore))
	http.Handle("/api/rooms", auth.BasicAuthMiddleware(userStore, http.HandlerFunc(api.HandleRooms(roomManager))))
	http.Handle("/ws", signaling.HandleWebSocket(roomManager, userStore))
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Println("SFU server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
