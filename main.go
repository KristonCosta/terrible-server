package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}


type server struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
}

func (server *server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	log.Printf("Client connected.")
	defer ws.Close()
	if server.clients == nil {
		server.clients = make(map[*websocket.Conn]bool)
	}
	server.clients[ws] = true
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Client Error: %v", err)
			delete(server.clients, ws)
			break
		}
		log.Printf("Client send message: %v", msg)
		server.broadcast <- msg
	}

}

func (server *server) handleMessages() {
	for {
		msg := <-server.broadcast
		log.Printf("Received message: %v", msg)
		for client := range server.clients {
			w, err := client.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error: %v", err)
				client.Close()
				delete(server.clients, client)
			}
			w.Write(msg)
			w.Close()
		}
	}
}

func main() {
	var serv server
	serv.broadcast = make(chan []byte)
	go serv.handleMessages()
	http.HandleFunc("/chat", serv.handleConnections)
	log.Fatal(http.ListenAndServe(":8123", nil))
}
