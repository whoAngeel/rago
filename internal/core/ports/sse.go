package ports

type SSEClient struct {
	ID      string
	UserID  int
	Channel chan SSEEvent
}

type SSEEvent struct {
	Type string
	Data any
}

type SSEManager interface {
	AddClient(userID int, client *SSEClient)
	RemoveClient(userID int, clientID string)
	SendToUser(userID int, event SSEEvent)
	SendToAll(event SSEEvent)
}
