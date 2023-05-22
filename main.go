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
	app := app.New(client.New(warshipServerAddr, clientTimeout))
	if err := app.Run(); err != nil {
		panic(err)
	}
}
