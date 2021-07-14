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

// Server struct is the core of this application
// It has 6 properties
// Mu -> sync.Mutex is to prevent concurrent writing to Clients
// Clients map[*Client]bool is a map of Client structs connected through sockets
// Operations are all the operations available to the clients (ex: ping, authenticate, publish)
// Operations have a code like 1, 2, ...
// RedisHandlers is a map of all RedisHandlers for specific RedisChannels that
// the application is subscribed to
type Server struct {
	Mu            sync.Mutex
	Clients       map[*Client]bool //
	RedisClient   *redis.Client
	Operations    map[string]func(client *Client, request Request) error
	Config        utils.Config
	RedisHandlers map[string]func(pubSubChan <-chan *redis.Message)
}

// Client is a struct for a single Client connected through sockets
// Mu sync.Mutex is to prevent concurrent writing to Ws *websocket.Conn
// Ws *websocket.Conn is a gorilla/websocket upgraded connection to the client
type Client struct {
	Mu sync.Mutex
	Ws *websocket.Conn
}

// CreateServer is returns a ready to use *Server
// Usage server := CreateServer()
func CreateServer() *Server {
	var err error
	server := &Server{}
	server.Clients = make(map[*Client]bool)
	server.Operations = make(map[string]func(client *Client, request Request) error)
	server.RedisHandlers = make(map[string]func(pubSubChan <-chan *redis.Message))
	server.Config, err = utils.LoadConfig("../", "app")
	if err != nil {
		log.Println(err)
	}
	server.prepareRedis()
	return server
}

// RegisterCommand is a method of s *Server
// It is used to safely add Operations *Server.Operations
// to the server.
// If there already exists a command with the operation code
// It returns an error
// Usage:
// err := s.RegisterCommand("1", Myfunc)
// if err != nil{
// 	log.Fatalln(err)
// }
func (s *Server) RegisterCommand(opCode string, callback func(client *Client, request Request) error) error {
	_, ok := s.Operations[opCode]
	if ok {
		return fmt.Errorf("a command with opCode %s already exists", opCode)
	}
	s.Operations[opCode] = callback

	return nil
}

// RegisterRedisHandler is a method of s *Server
// It is used to safely add Redis Handlers func(pubSubChan <- chan *redis.Message)
// to the server.
// If there already exists a command with the operation code
// It returns an error
// Usage:
// err := s.RegisterRedisHandler("myChannel", Myfunc)
// if err != nil{
// 	log.Fatalln(err)
// }
func (s *Server) RegisterRedisHandler(channelName string, handler func(pubSubChan <-chan *redis.Message)) error {
	_, ok := s.RedisHandlers[channelName]
	if ok {
		return fmt.Errorf("a handler for the redis channel: %s already exists", channelName)
	}
	s.RedisHandlers[channelName] = handler

	return nil
}

// UseCommand is used in HandleConnections
// When HandleConnections recieves an operation code from the client
// It uses UseCommand to return the corresponding function for the operation code
func (s *Server) UseCommand(opCode string) (func(client *Client, request Request) error, error) {
	callback, ok := s.Operations[opCode]
	if !ok {
		return nil, fmt.Errorf("no command with the opCode %s exists", opCode)
	}
	return callback, nil
}

// prepareRedis Prepares a Redis Connection
func (s *Server) prepareRedis() {
	s.RedisClient = redis.NewClient(
		&redis.Options{
			Addr:     s.Config.RedisHost,
			Password: s.Config.RedisPass,
			DB:       s.Config.RedisDb,
		})
}

func (s *Server) RegisterCommands() error {
	err := s.RegisterCommand("0", Ping)
	if err != nil {
		return err
	}

	err = s.RegisterCommand("1", PublishToRedis(s))
	if err != nil {
		return err
	}

	//err = s.RegisterCommand("2", Identify)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (s *Server) RegisterRedisHandlers() error {
	err := s.RegisterRedisHandler("challenges.new", NewChallengeSub(s))
	if err != nil {
		return err
	}
	return nil
}

// RunServer is used in main.go
// it prepares the server,
// adds registered redis handlers,
// adds registerd operations,
// and runs the server

func (s *Server) RunServer(host string, port int) (err error) {
	http.HandleFunc("/", HandleConnections(s))

	err = s.RegisterRedisHandlers()
	if err != nil {
		log.Fatal(err)
	}

	// Run Go Routine to handle Redis Events
	go RedisHandler(s)

	err = s.RegisterCommands()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting Server at %s:%v", host, port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%v", host, port), nil)
	return
}
