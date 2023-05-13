package app

import (
	"flag"
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
	RefreshSession()
	Fire() error
	CheckOpponentsDesc() (*StatusResponse, error)
	GetActivePlayersList() error
	WaitForValidOpponent() string
	ShowAccuracy()
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
	TargetNickPtr   *string
	NickPtr         *string
	DescPtr         *string
	Wpbot           *bool
	ActiveOpponents []string
)

func (a *app) Run() error {
	loadFlags()
	if !*Wpbot {
		go a.refreshList()
		a.c.WaitForValidOpponent()
	}
	if err := a.c.InitGame(); err != nil {
		return err
	}
	a.c.Board()
	a.checkStatus()
	return nil
}

func loadFlags() {
	TargetNickPtr = flag.String("target_nick", "", "Specify the target nickname")
	NickPtr = flag.String("nick", "", "Specify your nickname")
	DescPtr = flag.String("desc", "", "Specify your nickname")
	Wpbot = flag.Bool("wpbot", false, "Specify if you want to play with WP bot")
	flag.Parse()
}

func (a *app) refreshList() error {
	t := time.NewTicker(time.Second * 5)
	for range t.C {
		if err := a.c.GetActivePlayersList(); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) checkStatus() {
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
		a.c.RefreshSession()
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
