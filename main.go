package main

import (
	"log"
	"os"

	"github.com/Tech-With-Tim/Socket-Api/server"
	"github.com/urfave/cli/v2"
)

var app = cli.NewApp()

func main() {
	commands()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func commands() {
	app.Commands = []*cli.Command{
		{
			Name:  "runserver",
			Usage: "Run Api Server",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "host",
					Usage:   "Host on which server has to be run",
					Value:   "localhost",
					Aliases: []string{"H"},
				},
				&cli.IntFlag{
					Name:    "port",
					Usage:   "Port on which server has to be run",
					Value:   5000,
					Aliases: []string{"P"},
				},
			},
			Action: func(c *cli.Context) error {
				s := server.CreateServer()
				err := s.RegisterCommand("1", func(sender *server.Client, request server.Request) {
					for client := range s.Clients {
						if client.Ws != sender.Ws {
							client.Mu.Lock()
							err := client.Ws.WriteJSON(map[string]interface{}{"d": request.Data}) // like this
							client.Mu.Unlock()
							if err != nil {
								log.Println(err)
							}
						}
					}
				})
				if err != nil {
					return err
				}
				err = s.Runserver(c.String("host"), c.Int("port"))
				return err
			},
		},
	}
}

// func publish(s *server.Server) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ws, err := upgrader.Upgrade(w, r, nil)
// 		if err != nil {
// 			log.Println(err.Error())
// 		}
// 		defer ws.Close()
// 		s.Clients[ws] = true
// 		fmt.Println(s.Clients)
// 		for {
// 			var challenge Challenge
// 			err := ws.ReadJSON(&challenge)
// 			if err != nil {
// 				log.Printf("error: %v", err)
// 				delete(s.Clients, ws)
// 				break
// 			}
// 			msg, err := json.Marshal(challenge)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 			err = s.RedisClient.Publish(r.Context(), "challenges", msg).Err()
// 			if err != nil {
// 				log.Println(err)
// 			}

// 		}
// 	}

// }
