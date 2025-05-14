package api

import (
	"encoding/json"
	"net/http"

	"github.com/Coderovshik/meet/internal/auth"
	"github.com/Coderovshik/meet/internal/rooms"
)

func HandleRegister(us *auth.UserStore) http.HandlerFunc {
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
		w.WriteHeader(http.StatusCreated)
	}
}

func HandleLogin(us *auth.UserStore) http.HandlerFunc {
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleRooms(manager *rooms.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req struct{ ID string }
			_ = json.NewDecoder(r.Body).Decode(&req)

			usernameRaw := r.Context().Value(auth.UserContextKey)
			username, ok := usernameRaw.(string)
			if !ok || username == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !manager.CreateRoom(req.ID, username) {
				http.Error(w, "Room exists", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			var req struct{ ID string }
			_ = json.NewDecoder(r.Body).Decode(&req)
			manager.DeleteRoom(req.ID)
			w.WriteHeader(http.StatusOK)

		case http.MethodGet:
			// 🎯 Новый режим → /api/rooms?id=roomName
			roomID := r.URL.Query().Get("id")
			if roomID != "" {
				room, ok := manager.GetRoom(roomID)
				if !ok {
					http.Error(w, "Room not found", http.StatusNotFound)
					return
				}
				_ = json.NewEncoder(w).Encode(struct {
					Host    string `json:"host"`
					Creator string `json:"creator"`
				}{
					Host:    room.Host,
					Creator: room.Creator,
				})
				return
			}

			// обычный /api/rooms → список
			rooms := manager.ListRooms()
			_ = json.NewEncoder(w).Encode(rooms)

		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	}
}
