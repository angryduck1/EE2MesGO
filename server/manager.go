package server

import (
	"log"

	"github.com/gorilla/websocket"
)

const FixedQueueLimit = 500

var ResponseQueue chan interface{}
var semaphore = make(chan struct{}, FixedQueueLimit)

func clientWebSocketActivity(conn *websocket.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Conn close error: %v", err)
		}
	}()

	for {
		messageType, _, err := conn.NextReader()

		if err != nil {
			log.Printf("websocket error: %v", err)
			return
		}

		switch messageType {
		case 2:

		}
	}
}

func manageActivity() {
	for msg := range ResponseQueue {
		switch msg.(type) {

		case SyncInfo:
			defer func() {
				<-semaphore
			}()

			semaphore <- struct{}{}
		}
	}
}
