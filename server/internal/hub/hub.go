package hub

import "sync"

type sub struct {
	deviceID string
	ch       chan []byte
}

type Hub struct {
	mu     sync.RWMutex
	rooms  map[string]map[string]*sub
	bufCap int
}

func New(bufCap int) *Hub {
	if bufCap <= 0 {
		bufCap = 32
	}
	return &Hub{
		rooms:  make(map[string]map[string]*sub),
		bufCap: bufCap,
	}
}

func (h *Hub) Join(userID, deviceID string) (<-chan []byte, func()) {
	ch := make(chan []byte, h.bufCap)
	s := &sub{deviceID: deviceID, ch: ch}

	h.mu.Lock()
	if _, ok := h.rooms[userID]; !ok {
		h.rooms[userID] = make(map[string]*sub)
	}
	h.rooms[userID][deviceID] = s
	h.mu.Unlock()

	leave := func() {
		h.mu.Lock()
		if m, ok := h.rooms[userID]; ok {
			if m[deviceID] == s {
				delete(m, deviceID)
			}
			if len(m) == 0 {
				delete(h.rooms, userID)
			}
		}
		h.mu.Unlock()
		close(ch)
	}
	return ch, leave
}

func (h *Hub) Broadcast(userID, fromDevice string, payload []byte) {
	h.mu.RLock()
	m, ok := h.rooms[userID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	for id, s := range m {
		if id == fromDevice {
			continue
		}
		select {
		case s.ch <- payload:
		default:
		}
	}
	h.mu.RUnlock()
}
