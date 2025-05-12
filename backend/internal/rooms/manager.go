package rooms

import "sync"

type Manager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

func (m *Manager) CreateRoom(id, creator string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rooms[id]; exists {
		return false
	}
	m.rooms[id] = &Room{
		ID:      id,
		Creator: creator,
		Host:    creator,
		Peers:   make(map[string]*PeerConnection),
	}
	return true
}

func (m *Manager) DeleteRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

func (m *Manager) GetRoom(id string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room, exists := m.rooms[id]
	return room, exists
}

func (m *Manager) ListRooms() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.rooms))
	for id := range m.rooms {
		ids = append(ids, id)
	}
	return ids
}
