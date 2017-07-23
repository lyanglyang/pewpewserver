package main

import (
	"log"
	"net/http"
	"time"
	"github.com/googollee/go-socket.io"
	"strings"
	"fmt"
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

	//HTTP settings
	fmt.Printf("%slistening on port 5000%s \n", green, reset)
	http.Handle("/socket.io/", wsServer)
	http.ListenAndServe(":5000", nil)
}

func configureSocketIO() *socketio.Server {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	//Client connects to server
	server.On("connection", func(so socketio.Socket) {

		//What will happen as soon as the connection is established:
		so.On("connection", func(msg string) {
			so.Join("clients")
			println(so.Id() + " joined clients.")

			//In case you want to send a custom emit directly after the client connected.
			//If you fire an emit directly after the connection event it won't work therefore you need to wait a bit
			//In this case two seconds.
			ticker := time.NewTicker(2 * time.Second)
			go func() {
				for {
					select {
					case <-ticker.C:
						so.Emit("online", "Do Something!")
						ticker.Stop()
						return
					}
				}
			}()
		})

		//game events
		so.On("player-update", func(msg string) {
			fmt.Println("player-update", msg)
		})

		so.On("player-use-sword", func(msg string) {
			fmt.Println("player-use-sword", msg)
		})

		so.On("player-hit", func(msg string) {
			fmt.Println("player-hit", msg)
		})

		//What will happen if clients disconnect
		so.On("disconnection", func() {
			log.Println("on disconnect")
		})

		//Custom event as example
		so.On("hello", func(msg string) {
			log.Println("received request (hello): " + msg)

			so.Emit("Hi", "How can I help you?")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	return server
}
