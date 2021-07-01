package server

import (
	"log"
)

func Ping(client *Client, request Request) {
	client.Mu.Lock()
	err := client.Ws.WriteJSON(map[string]interface{}{"op": "0"})
	client.Mu.Unlock()
	if err != nil {
		log.Println(err)
	}
}
