package poker

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type PlayerServer struct {
	store PlayerStore
	http.Handler
	template *template.Template
	playGame Game
}

const htmlTemplatePath = "game.html"

const jsonContentType = "application/json"

type Player struct {
	Name string
	Wins int
}

func NewPlayerServer(store PlayerStore, game Game) (*PlayerServer, error) {
	p := new(PlayerServer)

	tmpl, err := template.ParseFiles(htmlTemplatePath)

	if err != nil {
		return nil, fmt.Errorf("problem opening %s %v", htmlTemplatePath, err)
	}

	p.playGame = game
	p.template = tmpl
	p.store = store

	router := http.NewServeMux()
	router.Handle("/league", http.HandlerFunc(p.leagueHandler))
	router.Handle("/players/", http.HandlerFunc(p.playerHandler))
	router.Handle("/game", http.HandlerFunc(p.game))
	router.Handle("/ws", http.HandlerFunc(p.webSocket))

	p.Handler = router

	return p, nil
}

type PlayerStore interface {
	GetPlayerScore(name string) int
	RecordWin(name string)
	GetLeague() League
}

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(p.store.GetLeague())
}

func (p *PlayerServer) playerHandler(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	switch r.Method {
	case http.MethodPost:
		p.processWin(w, player)
	case http.MethodGet:
		p.showScore(w, player)
	}
}

func (p *PlayerServer) game(w http.ResponseWriter, r *http.Request) {
	p.template.Execute(w, nil)
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type playerServerWS struct {
	*websocket.Conn
	writeMu sync.Mutex
}

func newPlayerServerWS(w http.ResponseWriter, r *http.Request) (*playerServerWS, error) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, fmt.Errorf("problem upgrading connection to WebSockets %v", err)
	}

	return &playerServerWS{Conn: conn}, nil
}

func (w *playerServerWS) WaitForMsg() (string, error) {
	_, msg, err := w.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("error reading from websocket %v", err)
	}
	return string(msg), nil
}

func (w *playerServerWS) Write(p []byte) (n int, err error) {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	err = w.WriteMessage(websocket.TextMessage, p)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (p *PlayerServer) webSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := newPlayerServerWS(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	numberOfPlayersMsg, err := ws.WaitForMsg()
	if err != nil {
		log.Println(err)
		return
	}

	numberOfPlayers, err := parseNumberOfPlayers(numberOfPlayersMsg)
	if err != nil {
		log.Println(err)
		return
	}

	stop := p.playGame.Start(numberOfPlayers, ws)
	defer stop()

	winnerMsg, err := ws.WaitForMsg()
	if err != nil {
		log.Println(err)
		return
	}

	winner := strings.TrimSpace(winnerMsg)
	if winner == "" {
		log.Println("no winner was sent over websocket, so the result was not recorded")
		return
	}

	p.playGame.Finish(winner)
}

func (p *PlayerServer) showScore(w http.ResponseWriter, player string) {
	score := p.store.GetPlayerScore(player)

	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter, player string) {
	p.store.RecordWin(player)
	w.WriteHeader(http.StatusAccepted)
}
