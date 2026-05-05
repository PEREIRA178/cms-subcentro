package ws

import (
	"log"
	"time"

	"cms-plazareal/internal/realtime"

	"github.com/gofiber/websocket/v2"
)

// DeviceSocket handles WebSocket connections from display devices (pantallas/totems)
func DeviceSocket(hub *realtime.Hub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		code := c.Params("code")
		if code == "" {
			log.Println("❌ WS device: missing code")
			return
		}

		client := &realtime.Client{
			Conn:       c,
			Type:       realtime.ClientDevice,
			DeviceCode: code,
			Send:       make(chan []byte, 64),
		}

		hub.Register(client)
		defer hub.Unregister(client)

		// Writer goroutine
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case msg, ok := <-client.Send:
					if !ok {
						c.WriteMessage(websocket.CloseMessage, nil)
						return
					}
					if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
						return
					}
				case <-ticker.C:
					// Ping to keep connection alive
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				}
			}
		}()

		// Reader loop (heartbeats from device)
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
			// Device heartbeat received — could update last_seen in DB
		}
	}
}

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
