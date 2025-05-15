package signaling

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Coderovshik/meet/internal/auth"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return strings.Contains(origin, "amogus.root-hub.ru")
	},
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

func HandleWebSocket(userStore *auth.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		password := r.URL.Query().Get("password")

		if username == "" || password == "" {
			http.Error(w, "Missing credentials or room", http.StatusBadRequest)
			return
		}

		valid, err := userStore.ValidateUser(r.Context(), username, password)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}
		c := &threadSafeWriter{conn, sync.Mutex{}}
		defer c.Close()

		peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		if err != nil {
			log.Printf("Failed to creates a PeerConnection: %v", err)

			return
		}
		defer peerConnection.Close()

		for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
			if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionRecvonly,
			}); err != nil {
				log.Printf("Failed to add transceiver: %v", err)

				return
			}
		}

		listLock.Lock()
		peerConnections = append(peerConnections, peerConnectionState{peerConnection, c})
		listLock.Unlock()

		peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i == nil {
				return
			}

			candidateString, err := json.Marshal(i.ToJSON())
			if err != nil {
				log.Printf("Failed to marshal candidate to json: %v", err)

				return
			}

			log.Printf("Send candidate to client: %s", candidateString)

			if writeErr := c.WriteJSON(&websocketMessage{
				Event: "candidate",
				Data:  string(candidateString),
			}); writeErr != nil {
				log.Printf("Failed to write JSON: %v", writeErr)
			}
		})

		peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
			log.Printf("Connection state change: %s", p)

			switch p {
			case webrtc.PeerConnectionStateFailed:
				if err := peerConnection.Close(); err != nil {
					log.Printf("Failed to close PeerConnection: %v", err)
				}
			case webrtc.PeerConnectionStateClosed:
				signalPeerConnections()
			default:
			}
		})

		peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
			log.Printf("Got remote track: Kind=%s, ID=%s, PayloadType=%d", t.Kind(), t.ID(), t.PayloadType())

			trackLocal := addTrack(t)
			defer removeTrack(trackLocal)

			buf := make([]byte, 1500)
			rtpPkt := &rtp.Packet{}

			for {
				i, _, err := t.Read(buf)
				if err != nil {
					return
				}

				if err = rtpPkt.Unmarshal(buf[:i]); err != nil {
					log.Printf("Failed to unmarshal incoming RTP packet: %v", err)

					return
				}

				rtpPkt.Extension = false
				rtpPkt.Extensions = nil

				if err = trackLocal.WriteRTP(rtpPkt); err != nil {
					return
				}
			}
		})

		peerConnection.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
			log.Printf("ICE connection state changed: %s", is)
		})

		signalPeerConnections()

		message := &websocketMessage{}
		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				log.Printf("Failed to read message: %v", err)

				return
			}

			log.Printf("Got message: %s", raw)

			if err := json.Unmarshal(raw, &message); err != nil {
				log.Printf("Failed to unmarshal json to message: %v", err)

				return
			}

			switch message.Event {
			case "candidate":
				candidate := webrtc.ICECandidateInit{}
				if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
					log.Printf("Failed to unmarshal json to candidate: %v", err)

					return
				}

				log.Printf("Got candidate: %v", candidate)

				if err := peerConnection.AddICECandidate(candidate); err != nil {
					log.Printf("Failed to add ICE candidate: %v", err)

					return
				}
			case "answer":
				answer := webrtc.SessionDescription{}
				if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
					log.Printf("Failed to unmarshal json to answer: %v", err)

					return
				}

				log.Printf("Got answer: %v", answer)

				if err := peerConnection.SetRemoteDescription(answer); err != nil {
					log.Printf("Failed to set remote description: %v", err)

					return
				}
			default:
				log.Printf("unknown message: %+v", message)
			}
		}
	}
}
