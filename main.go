package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}
var signingKey = []byte("SUPER_SECRET_KEY")

type Server struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
}

func (server *Server) fetchToken(w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Second * 20).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims
	tokenString, _ := token.SignedString(signingKey)
	w.Write([]byte(tokenString))
}

func (server *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
	if err == nil {
		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid.")
			return
		}
	} else {
		fmt.Fprint(w, "Invalid request.")
		return
	}
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

func (server *Server) handleMessages() {
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
	var serv Server
	serv.broadcast = make(chan []byte)
	go serv.handleMessages()
	http.HandleFunc("/chat", serv.handleConnections)
	http.HandleFunc("/token", serv.fetchToken)
	log.Fatal(http.ListenAndServe(":8123", nil))
}
