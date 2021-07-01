package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Tech-With-Tim/Socket-Api/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type Server struct {
	Mu          sync.Mutex
	Clients     map[*Client]bool //
	RedisClient *redis.Client
	Operations  map[string]func(client *Client, request Request) error
	Config      utils.Config
}

type Client struct {
	Mu sync.Mutex
	Ws *websocket.Conn
}

func CreateServer() *Server {
	var err error
	server := &Server{}
	server.Clients = make(map[*Client]bool)
	server.Operations = make(map[string]func(client *Client, request Request) error)
	server.Config, err = utils.LoadConfig("../", "app")
	if err != nil {
		log.Println(err)
	}
	server.prepareRedis()
	err = server.RegisterCommand("0", Ping)
	if err != nil {
		log.Fatal(err)
	}
	return server
}

func (s *Server) RegisterCommand(opCode string, callback func(client *Client, request Request) error) error {
	_, ok := s.Operations[opCode]
	if ok {
		return fmt.Errorf("a command with opCode %s already exists", opCode)
	}
	s.Operations[opCode] = callback

	return nil
}

func (s *Server) UseCommand(opCode string) (func(client *Client, request Request) error, error) {
	callback, ok := s.Operations[opCode]
	if !ok {
		return nil, fmt.Errorf("no command with the opCode %s exists", opCode)
	}
	return callback, nil
}

func (s *Server) prepareRedis() {
	s.RedisClient = redis.NewClient(
		&redis.Options{
			Addr:     s.Config.RedisHost,
			Password: s.Config.RedisPass,
			DB:       s.Config.RedisDb,
		})
}

func (s *Server) Runserver(host string, port int) (err error) {
	http.HandleFunc("/", HandleConnections(s))
	// go HandleChallenges(s)

	go Handler(s, map[string]func(message *redis.Message){

		"challenges.new": func(message *redis.Message) {
			fmt.Println("we are in")
			var challenge interface{} // here this is map
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
		},
	})
	log.Printf("Starting Server at %s:%v", host, port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%v", host, port), nil)
	return
}
