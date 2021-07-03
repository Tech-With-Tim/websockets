package server

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
)

func NewChallengeSub(s *Server) func(pubSubChan <- chan *redis.Message){
	return func(pubSubChan <- chan *redis.Message) {
		var challenge interface{}
		for {
			message := <-pubSubChan
			err := json.Unmarshal([]byte(message.Payload), &challenge)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			s.Mu.Lock()
			for client := range s.Clients {
				client.Mu.Lock()
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
