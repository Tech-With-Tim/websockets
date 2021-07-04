package server

import (
	"context"
	"encoding/json"
	"fmt"
)

func Ping(client *Client, request Request) error {
	client.Mu.Lock()
	err := client.Ws.WriteJSON(map[string]interface{}{"op": "0"})
	client.Mu.Unlock()
	if err != nil {
		return err
	}
	return nil
}

//func Identify(client *Client, request Request) error {
//	_, ok := (request.Data.(map[string]interface{}))["token"]
//	if !ok {
//		return fmt.Errorf("no token found")
//	}
//	return nil
//}

func PublishToRedis(s *Server) func(client *Client, request Request) error {
	return func(client *Client, request Request) error {
		msg, err := json.Marshal(request.Data)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		s.RedisClient.Publish(context.Background(), "challenges.new", msg)
		return nil
	}
}