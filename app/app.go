package app

import (
	"fmt"
	"log"
	"time"

	"main.go/flags"
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
	c                client
	GoroutineStopper chan bool
}

func New(c client) *app {
	return &app{
		c,
		make(chan bool, 1),
	}
}

var (
	ActiveOpponents []string
)

func (a *app) Run() error {
	if !*flags.WpbotFlag {
		go a.refreshList(a.GoroutineStopper)
		a.c.WaitForValidOpponent()
	}

	if err := a.c.InitGame(); err != nil {
		return err
	}
	a.c.Board()
	a.checkStatus()
	return nil
}

func (a *app) refreshList(stopper <-chan bool) error {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	fmt.Println("Refreshing list of active players")

	if err := a.c.GetActivePlayersList(); err != nil {
		log.Println("Error while refreshing list of active players")
		return err
	}

	for {
		select {
		case <-stopper:
			log.Println("Stopped refreshing list of active players")
			return nil
		case <-t.C:
			if err := a.c.GetActivePlayersList(); err != nil {
				fmt.Println("Error while refreshing list of active players")
				return err
			}
			fmt.Print("Type opponent's name: ")
		}
	}
}

func (a *app) checkStatus() {
	showInfo := true
	go a.refreshToken(a.GoroutineStopper)
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
			a.StopGoRoutines()
			return
		}
	}
}

func (a *app) refreshToken(stopper <-chan bool) {
	t := time.NewTicker(time.Second * 10)

	select {
	case <-stopper:
		t.Stop()
		return
	default:
		for range t.C {
			a.c.RefreshSession()
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

func (a *app) StopGoRoutines() {
	a.GoroutineStopper <- true
}
