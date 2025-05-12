package signaling

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Coderovshik/meet/internal/auth"
	"github.com/Coderovshik/meet/internal/rooms"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

var upgrader = websocket.Upgrader{}

func HandleWebSocket(manager *rooms.Manager, userStore *auth.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Query().Get("room")
		username := r.URL.Query().Get("username")
		password := r.URL.Query().Get("password")

		if roomID == "" || username == "" || password == "" {
			http.Error(w, "Missing credentials or room", http.StatusBadRequest)
			return
		}

		valid, _ := userStore.ValidateUser(r.Context(), username, password)
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// сохранить username в context
		r = r.WithContext(context.WithValue(r.Context(), auth.UserContextKey, username))

		room, ok := manager.GetRoom(roomID)
		if !ok {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		peerConnection, _ := webrtc.NewPeerConnection(webrtc.Configuration{})
		room.AddPeer(username, peerConnection)

		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			room.Mu.Lock()
			isHost := (username == room.Host)
			room.Mu.Unlock()

			if !isHost {
				return
			}

			room.Mu.Lock()
			defer room.Mu.Unlock()
			for id, pc := range room.Peers {
				if id == room.Host {
					continue
				}
				localTrack, _ := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())
				_, _ = pc.AddTrack(localTrack)
				go func() {
					buf := make([]byte, 1500)
					for {
						n, _, err := track.Read(buf)
						if err != nil {
							break
						}
						_, _ = localTrack.Write(buf[:n])
					}
				}()
			}
		})

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var payload map[string]interface{}
			_ = json.Unmarshal(msg, &payload)
			switch payload["type"] {
			case "offer":
				offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: payload["sdp"].(string)}
				_ = peerConnection.SetRemoteDescription(offer)
				answer, _ := peerConnection.CreateAnswer(nil)
				_ = peerConnection.SetLocalDescription(answer)
				resp, _ := json.Marshal(map[string]interface{}{"type": "answer", "sdp": answer.SDP})
				_ = conn.WriteMessage(websocket.TextMessage, resp)
			case "candidate":
				_ = peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: payload["candidate"].(string)})
			}
		}

		room.RemovePeer(username)
	}
}
