package main

import (
	"time"

	"main.go/app"
	"main.go/client"
)

const (
	warshipServerAddr = "https://go-pjatk-server.fly.dev"
	clientTimeout     = time.Second * 2
)

func main() {

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
