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
				err := s.RunServer(c.String("host"), c.Int("port"))
				return err
			},
		},
	}
}
