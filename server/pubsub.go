package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)
func NewChallengeSub(s *Server) func(message *redis.Message){
	return func (message *redis.Message) {
		var challenge interface{}
		err := json.Unmarshal([]byte(message.Payload), &challenge)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		fmt.Println(challenge)
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
