package rooms

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

type PeerConnection = webrtc.PeerConnection

type Room struct {
	ID      string
	Peers   map[string]*PeerConnection
	Host    string
	Creator string
	Mu      sync.Mutex
}

func (r *Room) AddPeer(id string, pc *PeerConnection) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Peers[id] = pc
}

func (r *Room) RemovePeer(id string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.Peers, id)
	if id == r.Host {
		r.Host = "" // опционально можно сделать автоназначение нового host
	}
}
