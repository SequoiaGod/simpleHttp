package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var peopleMap = make(map[string]int)
var messageChan = make(chan message)

type message struct {
	command  string
	key      string
	value    int
	response chan<- string
}

func run() {
	for mes := range messageChan {
		switch mes.command {
		case "get":
			var str strings.Builder
			str.WriteString("{")
			for k, v := range peopleMap {
				str.WriteString("\n" + k + ": " + strconv.Itoa(v))
			}
			str.WriteString("\n}")
			mes.response <- str.String()
		case "put":
			peopleMap[mes.key] = mes.value
			mes.response <- fmt.Sprintf("%s:%d", mes.key, mes.value)
		case "delete":
			delete(peopleMap, mes.key)
			mes.response <- fmt.Sprintf("%s has deleted", mes.key)
		}
	}
}

type People struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func startServer(ctx context.Context, server *http.Server) error {
	var err error

	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	<-ctx.Done()
	log.Println("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()
	if err = server.Shutdown(ctxShutDown); err != nil {
		log.Println(err)
	}

	if err == http.ErrServerClosed {
		err = nil
	}
	return err

}

func NewServer() {

	go run()
	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			func(w http.ResponseWriter, r *http.Request) {
				response := make(chan string)
				messageChan <- message{command: "get", response: response}
				fmt.Fprintln(w, <-response)
			}(w, r)
		case "POST":
			func(w http.ResponseWriter, r *http.Request) {
				people := requestDecoder(w, r)
				response := make(chan string)
				messageChan <- message{command: "put", key: people.Name, value: people.Age, response: response}
				fmt.Fprintln(w, <-response)
			}(w, r)
		case "PUT":
			func(w http.ResponseWriter, r *http.Request) {
				people := requestDecoder(w, r)
				response := make(chan string)
				messageChan <- message{command: "put", key: people.Name, value: people.Age, response: response}
				fmt.Fprintln(w, <-response)
			}(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/delete/"):]
		if name == "" {
			http.Error(w, "empty string", http.StatusBadRequest)
		}
		res := make(chan string)
		messageChan <- message{command: "delete", key: name, response: res}
		fmt.Fprintln(w, <-res)
	})
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	if err := startServer(ctx, server); err != nil {
		log.Println(err)
	}

	go func() {
		<-terminate
		log.Println("system closing---")
		cancel()
	}()

}
func requestDecoder(w http.ResponseWriter, r *http.Request) People {
	decoder := json.NewDecoder(r.Body)
	var people People
	err := decoder.Decode(&people)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return people
}
