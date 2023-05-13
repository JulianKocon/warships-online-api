package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	color "github.com/fatih/color"
	gui "github.com/grupawp/warships-lightgui/v2"
	"main.go/app"
)

type shots struct {
	hits   int
	misses int
}

type client struct {
	client     *http.Client
	serverAddr string
	token      string
	board      *gui.Board
	status     app.StatusResponse
	reader     bufio.Reader
	shots      shots
}

type waitingPlayer struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

func New(addr string, t time.Duration) *client {
	return &client{
		client: &http.Client{
			Timeout: t,
		},
		serverAddr: addr,
		reader:     *bufio.NewReader(os.Stdin),
	}
}

func (c *client) InitGame() error {
	wpbot := *app.Wpbot
	initBody := app.BasicRequestBody{
		Wpbot:      wpbot,
		TargetNick: *app.TargetNickPtr,
		Nick:       *app.NickPtr,
		Desc:       *app.DescPtr,
	}

	jsonBody, err := json.Marshal(initBody)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.serverAddr, "/api/game")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Print(string(body))
	}
	defer resp.Body.Close()
	c.token = resp.Header.Get("X-Auth-Token")
	return nil
}

func (c *client) Board() ([]string, error) {
	url, err := url.JoinPath(c.serverAddr, "/api/game/board")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var boardData map[string][]string
	if err := json.Unmarshal(body, &boardData); err != nil {
		return nil, err
	}
	setBoardConfig(c)

	c.board.Import(boardData["board"])
	c.board.Display()
	return c.board.Export(gui.Left), nil
}

func setBoardConfig(c *client) {
	cfg := gui.NewConfig()
	cfg.HitChar = '#'
	cfg.HitColor = color.FgRed
	cfg.BorderColor = color.BgRed
	cfg.RulerTextColor = color.BgYellow
	cfg.ShipColor = color.FgGreen
	c.board = gui.New(cfg)
}

func (c *client) Status() (*app.StatusResponse, error) {
	url, err := url.JoinPath(c.serverAddr, "/api/game")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gameStatusResponse := &app.StatusResponse{}
	if err := json.Unmarshal(body, gameStatusResponse); err != nil {
		return nil, err
	}
	c.status = *gameStatusResponse

	if len(c.status.OppShots) > 0 {
		for _, shot := range c.status.OppShots {
			c.board.HitOrMiss(gui.Left, shot)
		}
	}
	return gameStatusResponse, nil
}

func (c *client) UpdateBoard(status *app.StatusResponse) {
	c.board.Import(status.OppShots)
	c.board.Display()
}

func (c *client) RefreshSession() {
	url, err := url.JoinPath(c.serverAddr, "/api/game/refresh")
	if err != nil {
		return
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-Auth-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		log.Print(resp.StatusCode)
	}
	defer resp.Body.Close()

}
func (c *client) Fire() error {
	c.ShowAccuracy()
	fmt.Print("It's your turn:")
	input := waitForValidInput(c)

	reqBody := map[string]string{
		"coord": input,
	}
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	url, err := url.JoinPath(c.serverAddr, "/api/game/fire")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return err
	}

	req.Header.Set("X-Auth-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respMap map[string]string
	if err := json.Unmarshal([]byte(body), &respMap); err != nil {
		return err
	}

	if respMap["result"] == "hit" {
		c.shots.hits++
		c.board.Set(gui.Right, input, gui.Hit)
		c.board.Display()
		c.Fire()
	} else if respMap["result"] == "sunk" {
		c.shots.hits++
		c.board.Set(gui.Right, input, gui.Hit)
		c.board.CreateBorder(gui.Right, input)
		c.board.Display()
		c.Fire()
	} else {
		c.shots.misses++
		c.board.Set(gui.Right, input, gui.Miss)
		c.board.Display()
	}

	return nil
}

func waitForValidInput(c *client) string {
	input, _ := c.reader.ReadString('\n')
	input = strings.TrimSpace(input)
	pattern := regexp.MustCompile(`[A-J][1-9]|10`)
	if !pattern.MatchString(input) {
		fmt.Print(gui.ErrInvalidCoord, "\nType again:")
		waitForValidInput(c)
	}
	return input
}

func (c *client) CheckOpponentsDesc() (*app.StatusResponse, error) {
	url, err := url.JoinPath(c.serverAddr, "/api/game/desc")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gameStatusResponse := &app.StatusResponse{}
	if err := json.Unmarshal(body, gameStatusResponse); err != nil {
		return nil, err
	}
	c.status = *gameStatusResponse

	return gameStatusResponse, nil
}

func (c *client) GetActivePlayersList() error {

	url, err := url.JoinPath(c.serverAddr, "/api/game/list")
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var waitingPlayer []waitingPlayer
	if err := json.Unmarshal([]byte(body), &waitingPlayer); err != nil {
		return err
	}
	for _, opponent := range waitingPlayer {
		if opponent.GameStatus == "waiting" {
			opponent := opponent.Nick
			app.ActiveOpponents = append(app.ActiveOpponents, opponent)
		}
	}

	if waitingPlayer == nil {
		log.Println("No active players.")
	} else {
		log.Println("Active players:")
		for _, item := range waitingPlayer {
			log.Println(item.Nick)
		}
	}

	return nil
}

func (c *client) WaitForValidOpponent() string {
	input, _ := c.reader.ReadString('\n')
	for _, opponent := range app.ActiveOpponents {
		if opponent == input {
			return input
		}
	}
	log.Print("Invalid opponent. Type again:")
	return c.WaitForValidOpponent()
}

func (c *client) ShowAccuracy() {
	if c.shots.hits == 0 && c.shots.misses == 0 {
		log.Println("Accuracy: 0%")
		return
	}
	accuracy := float64(c.shots.hits) / float64(c.shots.hits+c.shots.misses) * 100
	log.Printf("Accuracy: %.2f%%", accuracy)
}
