package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Message struct {
	Id      string
	Content string
}

type Messages []Message

func get_messages(w http.ResponseWriter, r *http.Request) {
	var messages Messages

	db, err := gorm.Open("sqlite3", "./foo.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Find(&messages)
	json.NewEncoder(w).Encode(messages)
}

func save_message(w http.ResponseWriter, r *http.Request) {
	var message Message

	db, err := gorm.Open("sqlite3", "./foo.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&message)
	if err != nil {
		http.Error(w, "Invalid post body.", http.StatusBadRequest)
		return
	}
	message.Id = strconv.FormatInt(time.Now().UnixNano(), 10)
	db.Create(&message)
	w.Write([]byte("Success!"))
}

func main() {
	db, err := gorm.Open("sqlite3", "./foo.db")

	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.AutoMigrate(&Message{})

	router := mux.NewRouter().StrictSlash(true)
	router.
		Methods("POST").
		Path("/messages").
		HandlerFunc(save_message)

	router.
		Methods("GET").
		Path("/messages").
		HandlerFunc(get_messages)

	log.Fatal(http.ListenAndServe(":8123", router))
}
