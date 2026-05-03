package sse

import (
	"sync"

	"github.com/whoAngeel/rago/internal/core/ports"
)

type Manager struct {
	mu      sync.RWMutex
	clients map[int]map[string]*ports.SSEClient
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[int]map[string]*ports.SSEClient),
	}
}

func (m *Manager) AddClient(userID int, client *ports.SSEClient) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[userID] == nil {
		m.clients[userID] = make(map[string]*ports.SSEClient)
	}
	m.clients[userID][client.ID] = client
}

func (m *Manager) RemoveClient(userID int, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userClients, ok := m.clients[userID]
	if !ok {
		return
	}

	if client, exists := userClients[clientID]; exists {
		close(client.Channel)
		delete(userClients, clientID)
	}

	if len(userClients) == 0 {
		delete(m.clients, userID)
	}
}

func (m *Manager) SendToUser(userID int, event ports.SSEEvent) {
	m.mu.RLock()
	userClients := m.clients[userID]
	m.mu.RUnlock()

	for _, client := range userClients {
		go m.sendEvent(client, event)
	}
}

func (m *Manager) SendToAll(event ports.SSEEvent) {
	m.mu.RLock()
	for _, userClients := range m.clients {
		for _, client := range userClients {
			go m.sendEvent(client, event)
		}
	}
	m.mu.RUnlock()
}

func (m *Manager) sendEvent(client *ports.SSEClient, event ports.SSEEvent) {
	defer func() {
		recover()
	}()
	select {
	case client.Channel <- event:
	default:
	}
}
