package server

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
)

// NewChallengeSub subscribes to challenges.new
// channel in redis.
// It sends recieved values to all connected clients
func NewChallengeSub(s *Server) func(pubSubChan <-chan *redis.Message) {
	return func(pubSubChan <-chan *redis.Message) {
		var challenge interface{}
		for {
			// pubSubChan is go channel for the subscribed channel on redis
			message := <-pubSubChan
			err := json.Unmarshal([]byte(message.Payload), &challenge)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			s.Mu.Lock()
			for client := range s.Clients {
				client.Mu.Lock()
				// Write to client
				err := client.Ws.WriteJSON(challenge)
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
}
