package main

import (
	"context"
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
		ctx, cancel := context.WithCancel(context.Background())
		app := app.New(ctx, client.New(warshipServerAddr, clientTimeout))
		defer cancel()
		if err := app.Run(); err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 30)
	}
}
