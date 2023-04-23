package app

import (
	"fmt"
	"log"
	"time"
)

type StatusResponse struct {
	Desc           string   `json:"desc"`
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
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
	RefreshSession()
	Fire() error
}
type app struct {
	c client
}

func New(c client) *app {
	return &app{
		c,
	}
}
func (a *app) Run() {
	a.c.InitGame()
	a.c.Board()
	a.checkStatus()

}

func (a *app) checkStatus() {
	showInfo := true
	for {
		time.Sleep(1 * time.Second)
		resp, err := a.c.Status()
		if err != nil {
			log.Fatal(err)
		}
		if resp.GameStatus == "" {
			log.Fatal(err)
		} else if resp.GameStatus == "game_in_progress" {
			if resp.ShouldFire {
				a.showGameInfoOnce(resp, &showInfo)
				if err := a.c.Fire(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

}

func (a *app) showGameInfoOnce(resp *StatusResponse, showInfo *bool) {
	if *showInfo {
		a.c.UpdateBoard(resp)
		fmt.Printf("Nick: %v \n", resp.Nick)
		fmt.Printf("Description: %v \n", resp.Desc)
		fmt.Printf("Opponent: %v \n", resp.Opponent)
		fmt.Printf("Opponent's description : %v \n", resp.OppDesc)
		*showInfo = false
	}
}
