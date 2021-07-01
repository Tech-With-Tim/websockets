package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Tech-With-Tim/Socket-Api/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type Server struct {
	Mu 			sync.Mutex
	Clients     map[*Client]bool
	RedisClient *redis.Client
	Operations  map[string]func(client *Client, request Request)
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
	server.Operations = make(map[string]func(client *Client, request Request))
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

func (s *Server) RegisterCommand(opCode string, callback func(client *Client, request Request)) error {
	_, ok := s.Operations[opCode]
	if ok {
		return fmt.Errorf("a command with opCode %s already exists", opCode)
	}
	s.Operations[opCode] = callback

	return nil
}

func (s *Server) UseCommand(opCode string) (func(client *Client, request Request), error) {
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
	go HandleChallenges(s)

	log.Printf("Starting Server at %s:%v", host, port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%v", host, port), nil)
	return
}
