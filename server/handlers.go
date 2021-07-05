package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// HandleConnections returns a httpHandlerfunc
// for ever client that joins
// it upgrades the connection to a socket connection
// Registers the client in *Server.Clients
// Writes time elapsed pings to the clients
// Listens for Operation Messages from clients and executes them
func HandleConnections(s *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client := &Client{}
		var err error
		client.Ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err.Error())
		}

		open := time.Now()

		// At the end of the connection
		// remove client from *Server.Clients map
		defer func() {
			s.Mu.Lock()
			delete(s.Clients, client)
			s.Mu.Unlock()
			client.Ws.Close()
		}()
		// Lock Mutex and add Clients to *Server.Clients Map
		s.Mu.Lock()
		s.Clients[client] = true
		fmt.Println(s.Clients)
		s.Mu.Unlock()

		// Send Pings To Clients containing time elapsed
		go func() {
			for {
				client.Mu.Lock()
				err := client.Ws.WriteJSON(map[string]interface{}{
					"op": "0",
					"d": map[string]string{
						"t": fmt.Sprint(time.Since(open).Milliseconds()),
					},
				})
				client.Mu.Unlock()
				if err != nil {
					break
				}
				time.Sleep(20 * time.Second)
			}
		}()

		// Listens to Client Requests in Json Format
		for {
			var request Request
			// Set deadline for participation from client
			// if exceeds deadline, client connection is closed
			err := client.Ws.SetReadDeadline(time.Now().Add(3 * time.Minute))
			if err != nil {
				log.Printf("error: %v", err)
			}

			// Listen incoming Json Requests
			err = client.Ws.ReadJSON(&request)
			if err != nil {
				log.Printf("error: %v", err)
				break
			}

			// Get Callback function for the corresponding Operation Code
			// Sent by the client
			callback, err := s.UseCommand(request.OperationCode)
			if err != nil {
				continue
			}

			// Execute the callback function
			// If errors remove client
			go func() {
				err := callback(client, request)
				if err != nil {
					fmt.Println(err.Error())
					s.Mu.Lock()
					delete(s.Clients, client)
					s.Mu.Unlock()
					client.Ws.Close()
				}
			}()
		}
	}
}

// RedisHandler subscribes too all registered
// Redis Channels and runs the respective handlers in
// goroutines
func RedisHandler(s *Server) {
	// loop over the hanlders map
	for channel, handler := range s.RedisHandlers {
		ctx := context.Background()
		// subscribe to redis channel
		pubsub := s.RedisClient.Subscribe(ctx, channel)
		// try recieving once
		_, err := pubsub.Receive(ctx)
		if err != nil {
			panic(err)
		}
		redisChan := pubsub.Channel()
		// run handler
		go handler(redisChan)
	}
}
