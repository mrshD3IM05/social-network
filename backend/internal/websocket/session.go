package websocket

import "sync"

var sessionClients = struct {
	sync.Mutex
	clients map[string]map[*Client]struct{}
}{clients: make(map[string]map[*Client]struct{})}

var clientSessions sync.Map

func trackClient(sessionID string, client *Client) {
	sessionClients.Lock()
	if sessionClients.clients[sessionID] == nil {
		sessionClients.clients[sessionID] = make(map[*Client]struct{})
	}
	sessionClients.clients[sessionID][client] = struct{}{}
	sessionClients.Unlock()
	clientSessions.Store(client, sessionID)
}

func untrackClient(client *Client) {
	value, ok := clientSessions.LoadAndDelete(client)
	if !ok {
		return
	}
	sessionID := value.(string)
	sessionClients.Lock()
	if clients := sessionClients.clients[sessionID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(sessionClients.clients, sessionID)
		}
	}
	sessionClients.Unlock()
}

func (h *Hub) RevokeSessionClients(sessionID string) {
	sessionClients.Lock()
	clients := sessionClients.clients[sessionID]
	delete(sessionClients.clients, sessionID)
	sessionClients.Unlock()
	for client := range clients {
		clientSessions.Delete(client)
		_ = client.connection.Close()
	}
}
