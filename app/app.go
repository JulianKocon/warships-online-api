package app

import (
	"fmt"
	"log"
	"os"
	"time"
)

type StatusResponse struct {
	Desc           string   `json:"desc"`
	GameStatus     string   `json:"game_status"`      //ended
	LastGameStatus string   `json:"last_game_status"` //win //lose
	Nick           string   `json:"nick"`
	OppDesc        string   `json:"opp_desc"`
	OppShots       []string `json:"opp_shots"`
	Opponent       string   `json:"opponent"`
	ShouldFire     bool     `json:"should_fire"`
	Timer          int      `json:"timer"`
}
type BasicRequestBody struct {
	Cords      []string `json:"cords,omitempty"`
	Desc       string   `json:"desc,omitempty"`
	Nick       string   `json:"nick,omitempty"`
	TargetNick string   `json:"target_nick,omitempty"`
	Wpbot      bool     `json:"wpbot,omitempty"`
}

type client interface {
	InitGame() error
	Board() ([]string, error)
	Status() (*StatusResponse, error)
	UpdateBoard(*StatusResponse)
	RefreshSession() error
	Fire() error
	CheckOpponentsDesc() (*StatusResponse, error)
	Abandon() error
}

type app struct {
	c client
}

func New(c client) *app {
	return &app{
		c,
	}
}

var (
	Abandon *bool
)

func (a *app) Run() {
	a.c.InitGame()
	_, err := a.c.Board()
	if err != nil {
		log.Fatal(err)
	}
	a.checkStatus()

}

func (a *app) checkStatus() {
	timer := StartWaitingRoomTimer()
	showInfo := true
	go a.refreshToken()
	for {
		time.Sleep(1 * time.Second)
		resp, err := a.c.Status()
		if err != nil {
			log.Fatal(err)
		}
		switch resp.GameStatus {
		case "game_in_progress":
			{
				timer.Stop()
				log.Print("Game in progress")
				if *Abandon {
					log.Print("Abandon")
					if err := a.c.Abandon(); err != nil {
						log.Fatal(err)
					}
				}
				if resp.ShouldFire {
					a.showGameInfoOnce(resp, &showInfo)
					if err := a.c.Fire(); err != nil {
						log.Fatal(err)
					}
				}
			}
		case "ended":
			fmt.Print("You ", resp.LastGameStatus, "!!!")
			os.Exit(1)
		}
	}
}

func (a *app) refreshToken() {
	t := time.NewTicker(time.Second * 10)
	for range t.C {
		if err := a.c.RefreshSession(); err != nil {
			log.Fatal(err)
		}
	}
}

func (a *app) showGameInfoOnce(resp *StatusResponse, showInfo *bool) {
	if *showInfo {
		resp, err := a.c.CheckOpponentsDesc()
		if err != nil {
			log.Fatal(err)
		}
		a.c.UpdateBoard(resp)
		fmt.Printf("Nick: %v \n", resp.Nick)
		fmt.Printf("Description: %v \n", resp.Desc)
		fmt.Printf("Opponent: %v \n", resp.Opponent)
		fmt.Printf("Opponent's description : %v \n", resp.OppDesc)
		*showInfo = false
	}
}

func StartWaitingRoomTimer() *time.Timer {
	return time.AfterFunc(3*time.Second, func() {
		log.Print("No activity for 3 seconds. Exiting program.")
		os.Exit(1)
	})
}
