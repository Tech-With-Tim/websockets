package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func HandleConnections(s *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client := &Client{}
		var err error
		client.Ws, err = upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println(err.Error())
		}

		open := time.Now()

		defer func() {
			s.Mu.Lock()
			delete(s.Clients, client)
			s.Mu.Unlock()
			client.Ws.Close()
		}()
		s.Mu.Lock()
		s.Clients[client] = true
		fmt.Println(s.Clients)
		s.Mu.Unlock()
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
				time.Sleep(5 * time.Second)
			}
		}()

		for {
			var request Request
			err := client.Ws.SetReadDeadline(time.Now().Add(3 * time.Minute))
			if err != nil {
				log.Printf("er ror: %v", err)
			}
			err = client.Ws.ReadJSON(&request)
			if err != nil {
				log.Printf("error: %v", err)
				break
			}

			callback, err := s.UseCommand(request.OperationCode)
			if err != nil {
				continue
			}

			go func() {
				err := callback(client, request)
				if err != nil {
					s.Mu.Lock()
					delete(s.Clients, client)
					s.Mu.Unlock()
					client.Ws.Close()
				}
			}()
		}
	}
}

func Handler(s *Server, handlers map[string]func(message *redis.Message)) {
	for channel, handler := range handlers {
		ctx := context.Background()
		pubsub := s.RedisClient.Subscribe(ctx, channel)
		_, err := pubsub.Receive(ctx)
		if err != nil {
			panic(err)
		} // try subscribe to channel from cli, check if messages get published
		redisChan := pubsub.Channel()
		go func(handler func(message *redis.Message)) {
			for {
				msg := <-redisChan
				handler(msg)
			}
		}(handler)
	}
}

// need better handlers
func HandleChallenges(s *Server) {
	var ctx = context.Background()

	pubsub := s.RedisClient.Subscribe(ctx, "events")
	_, err := pubsub.Receive(ctx)
	if err != nil {
		panic(err)
	}

	redisChan := pubsub.Channel()
	var subEvent SubEvent

	for {
		msg := <-redisChan
		err = json.Unmarshal([]byte(msg.Payload), &subEvent)
		if err != nil {
			log.Println(err)
		}
		s.Mu.Lock()
		for client := range s.Clients {
			client.Mu.Lock()
			err := client.Ws.WriteJSON(subEvent)
			client.Mu.Unlock()
			if err != nil {
				log.Printf("error: %v", err)
				delete(s.Clients, client)
				client.Ws.Close()
			}
		}
		s.Mu.Unlock()
	}
}
