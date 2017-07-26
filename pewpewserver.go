package main

import (
	"log"
	"net/http"
	"github.com/googollee/go-socket.io"
	"strings"
	"fmt"
	"os/exec"
	"os"
)

var green = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
var reset = string([]byte{27, 91, 48, 109})

//Custom server which basically only contains a socketio variable
//But we need it to enhance it with functions
type customServer struct {
	Server *socketio.Server
}

//Header handling, this is necessary to adjust security and/or header settings in general
//Please keep in mind to adjust that later on in a productive environment!
//Access-Control-Allow-Origin will be set to whoever will call the server
func (s *customServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allowedOrigins := "http://localhost:3000/, http://localhost:3000"
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")

	if strings.Contains(allowedOrigins, origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	s.Server.ServeHTTP(w, r)
}

func main() {
	//get/configure socket.io websocket for clients
	ioServer := configureSocketIO()

	wsServer := new(customServer)
	wsServer.Server = ioServer

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	//HTTP settings
	fmt.Printf("%slistening on port %s %s \n", green, port, reset)
	http.Handle("/socket.io/", wsServer)
	http.ListenAndServe(":"+port, nil)
}

type Gamer struct {
	Id       string `json:"id""`
	SocketId string `json:"socketId"`
	Name     string `json:"name""`
}

var gamers []Gamer

func configureSocketIO() *socketio.Server {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	//Client connects to server
	server.On("connection", func(so socketio.Socket) {

		so.Join("gamers")

		so.On("signup", func(name string) {
			out, err := exec.Command("uuidgen").Output()
			if err != nil {
				log.Fatal(err)
			}
			gamer := Gamer{Id: string(out), SocketId: so.Id(), Name: name}
			gamers = append(gamers, gamer)
			so.Emit("joined-successfully", gamer)
			so.BroadcastTo("gamers", "player-joined", gamer)
		})

		//game events
		so.On("player-update", func(msg string) {
			//fmt.Println("player-update", msg)
			so.BroadcastTo("gamers", "player-update", msg)
		})

		so.On("player-use-sword", func(msg string) {
			//fmt.Println("player-use-sword", msg)
			so.Emit("player-use-sword", msg)
			so.BroadcastTo("gamers", "player-use-sword", msg)
		})

		so.On("player-hit", func(msg string) {
			//fmt.Println("player-hit", msg)
			so.BroadcastTo("gamers", "player-hit", msg)
		})

		//What will happen if clients disconnect
		so.On("disconnection", func() {
			for i, v := range gamers {
				if v.SocketId == so.Id() {
					log.Println("on disconnect removed:", so.Id())
					gamers = append(gamers[:i], gamers[i+1:]...)
					break
				}
			}
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	return server
}
