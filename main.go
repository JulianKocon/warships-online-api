package main

import (
	"time"

	"main.go/app"
	"main.go/client"
	"main.go/flags"
)

const (
	warshipServerAddr = "https://go-pjatk-server.fly.dev"
	clientTimeout     = time.Second * 2
)

func main() {
	flags.LoadFlags()
	if err := flags.ValidateFlags(); err != nil {
		panic(err)
	}

	for {
		app := app.New(client.New(warshipServerAddr, clientTimeout))
		if err := app.Run(); err != nil {
			app.StopGoRoutines()
			panic(err)
		}
		app.StopGoRoutines()
		time.Sleep(time.Second * 30)
	}
}
