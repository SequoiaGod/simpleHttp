package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var messageChan = make(chan message)
var peopleMap = make(map[string]int)

type message struct {
	command  string
	key      string
	value    int
	response chan<- string
}

type People struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func getPeople(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getting people")
	response := make(chan string)
	messageChan <- message{command: "get", response: response}
	_, err := fmt.Fprintln(w, <-response)
	if err != nil {
		return
	}
}

func createPeople(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pareparing for creating people ...")
	people := requestDecoder(w, r)
	fmt.Println("createPeople")
	response := make(chan string)
	messageChan <- message{command: "put", key: people.Name, value: people.Age, response: response}
	fmt.Fprintln(w, <-response)
}

func deletePeopleByName(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	res := make(chan string)
	messageChan <- message{command: "delete", key: key, response: res}
	fmt.Fprintln(w, <-res)

}

func initializeRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /people", getPeople)
	mux.HandleFunc("POST /people", createPeople)
	mux.HandleFunc("DELETE /people/{key}", deletePeopleByName)
	mux.HandleFunc("PUT /people", createPeople)
	return mux
}

func NewServer() {
	ctx, cancel := context.WithCancel(context.Background())
	router := initializeRouter()
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	done := startServer(ctx, server)
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)

	<-terminate
	log.Println("system closing---")
	cancel()
	fmt.Println("closing")
	<-done

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

func run(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(messageChan)
	}()
	go func() {
		defer close(done)

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
				fmt.Println(mes.key, ": ", mes.value)
				mes.response <- fmt.Sprintf("%s:%d", mes.key, mes.value)
			case "delete":
				delete(peopleMap, mes.key)
				mes.response <- fmt.Sprintf("%s has deleted", mes.key)

			}
		}
	}()
	return done
}

func startServer(ctx context.Context, server *http.Server) <-chan struct{} {
	var err error
	done := make(chan struct{})
	runDone := run(ctx)
	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	<-ctx.Done()
	log.Println("server stopped")

	go func() {
		defer close(done)
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Println("Shutdown: " + err.Error())
		}
		<-runDone
	}()
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()
	if err = server.Shutdown(ctxShutDown); err != nil {
		log.Println(err)
	}

	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return done

}
