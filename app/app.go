package app

import (
	"errors"
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
	GetWaitingOpponents() ([]WaitingOpponent, error)
	GetOnlineOpponents() ([]WaitingOpponent, error)
	ShowAccuracy()
	Abandon()
	ReadInput() (string, error)
}

type app struct {
	c                         client
	GoroutineStopper          chan bool
	RefreshWaitingListStopper chan bool
	WaitingOpponents          []WaitingOpponent
}

func New(c client) *app {
	return &app{
		c,
		make(chan bool, 1),
		make(chan bool, 1),
		[]WaitingOpponent{},
	}
}

type WaitingOpponent struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

func (a *app) Run() error {
	flags.LoadFlags()
	if err := flags.ValidateFlags(); err != nil {
		return err
	}

	if !*flags.WpbotFlag && *flags.TargetNickFlag == "" && !*flags.WaitingFlag {
		go a.refreshWaitingPlayersList(a.RefreshWaitingListStopper)
		isOpponentValid, err := a.WaitForValidOpponent()
		if err != nil {
			return err
		}
		if isOpponentValid {
			a.RefreshWaitingListStopper <- true
		}
	}

	isNickAvailable, err := a.IsNickAvailable()
	if err != nil {
		return err
	}
	if !isNickAvailable {
		return errors.New("nick is not available")
	}

	if err := a.c.InitGame(); err != nil {
		return err
	}
	a.c.Board()
	a.checkStatus()
	return nil
}

func (a *app) refreshWaitingPlayersList(stopper <-chan bool) error {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	fmt.Println("Refreshing list of active players")

	if err := a.showWaitingOpponents(); err != nil {
		return err
	}

	for {
		select {
		case <-stopper:
			log.Println("Stopped refreshing list of active players")
			return nil
		case <-t.C:
			if err := a.showWaitingOpponents(); err != nil {
				return err
			}
		}
	}
}

func (a *app) showWaitingOpponents() (err error) {
	a.WaitingOpponents, err = a.c.GetWaitingOpponents()
	if err != nil {
		log.Println("Error while refreshing list of waiting opponents")
		return err
	}
	if len(a.WaitingOpponents) == 0 {
		fmt.Println("No active opponents")
	} else {
		fmt.Printf("Active opponents: \n")
		for _, opponent := range a.WaitingOpponents {
			fmt.Println(opponent.Nick)
		}
		fmt.Println("Type opponent's name: ")
	}
	return nil
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

		if resp.GameStatus == "game_in_progress" && resp.ShouldFire {
			a.showGameInfoOnce(resp, &showInfo)
			if err := a.c.Fire(); err != nil {
				if err.Error() == "abandon" {
					resp.LastGameStatus = "abandoned"
					resp.GameStatus = "ended"
				} else {
					log.Fatal(err)
				}
			}
		}

		if resp.GameStatus == "ended" {
			fmt.Print("You ", resp.LastGameStatus, "!!!")
			a.StopGoRoutines()
			fmt.Println("Restarting game in 30 seconds")
			return
		}
	}
}

func (a *app) refreshToken(stopper <-chan bool) {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-stopper:
			log.Println("Stopped refreshing token")
			return
		case <-t.C:
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

func (a *app) WaitForValidOpponent() (bool, error) {
	input, err := a.c.ReadInput()
	if err != nil {
		return false, errors.New("error while reading input")
	}
	for _, opponent := range a.WaitingOpponents {
		if opponent.GameStatus == "waiting" && opponent.Nick == input {
			return true, nil
		}
	}
	log.Print("Invalid opponent. Type again:")
	return a.WaitForValidOpponent()
}
func (a *app) IsNickAvailable() (bool, error) {
	for _, opponent := range a.WaitingOpponents {
		if opponent.Nick == *flags.NickFlag {
			return false, nil
		}
	}

	waitingOpponents, err := a.c.GetWaitingOpponents()
	a.WaitingOpponents = waitingOpponents
	if err != nil {
		return false, err
	}
	for _, player := range a.WaitingOpponents {
		if player.Nick == *flags.NickFlag && player.GameStatus == "waiting" {
			return false, nil
		}
	}
	return true, nil
}
