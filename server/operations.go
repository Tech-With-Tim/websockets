package server

import "fmt"

func Ping(client *Client, request Request) error {
	client.Mu.Lock()
	err := client.Ws.WriteJSON(map[string]interface{}{"op": "0"})
	client.Mu.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func Identify(client *Client, request Request) error {
	_, ok := (request.Data.(map[string]interface{}))["token"]
	if !ok {
		return fmt.Errorf("no token found")
	}
	return nil
}
