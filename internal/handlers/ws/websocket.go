package ws

import (
	"cms-plazareal/internal/realtime"

	"github.com/gofiber/websocket/v2"
)

// WebSocket handles WebSocket connections from the public website
func WebSocket(hub *realtime.Hub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		client := &realtime.Client{
			Conn: c,
			Type: realtime.ClientWeb,
			Send: make(chan []byte, 64),
		}

		hub.Register(client)
		defer hub.Unregister(client)

		// Writer goroutine
		go func() {
			for msg := range client.Send {
				if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}()

		// Reader loop
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
