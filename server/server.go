package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type MapStore struct {
	peopleMap map[string]int
	mu        sync.RWMutex
}

type People struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func NewServer() {
	mapStore := &MapStore{peopleMap: make(map[string]int)}
	server := &http.Server{Addr: ":8080"}
	mapStore.peopleMap["yan"] = 25
	mapStore.peopleMap["gin"] = 44
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getHandler(w, r, mapStore.peopleMap)
		case "POST":
			postHandler(w, r, mapStore.peopleMap)
		case "PUT":
			updateHandler(w, r, mapStore.peopleMap)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/test/"):]
		_, ok := mapStore.peopleMap[name]
		if !ok {
			http.Error(w, "people not found", http.StatusBadRequest)
		}
		fmt.Fprintln(w, name, ":", mapStore.peopleMap[name])
		delete(mapStore.peopleMap, name)

	})
	terminate := make(chan os.Signal, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	go func() {
		<-terminate
		err := server.Close()
		if err != nil {
			log.Fatal("server error: ", err)
		}
	}()

	<-terminate
}

func updateHandler(w http.ResponseWriter, r *http.Request, peopleMap map[string]int) {
	people := requestDecoder(w, r)
	peopleMap[people.Name] = people.Age
	fmt.Fprintln(w, people.Name, ":", people.Age)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, peopleMap map[string]int) {
	name := r.URL.Path[len("/test/"):]
	_, ok := peopleMap[name]
	if !ok {
		http.Error(w, "people not found", http.StatusBadRequest)
	}
	fmt.Fprintln(w, name, ":", peopleMap[name])
	delete(peopleMap, name)

}

func postHandler(w http.ResponseWriter, r *http.Request, peopleMap map[string]int) {
	people := requestDecoder(w, r)
	peopleMap[people.Name] = people.Age
	fmt.Fprintln(w, people.Name, ":", people.Age)
}

func getHandler(w http.ResponseWriter, r *http.Request, peopleMap map[string]int) {
	fmt.Fprintln(w, "{")
	for k, v := range peopleMap {
		fmt.Fprintf(w, "  %v: %v,\n", k, v)
	}
	fmt.Fprintln(w, "}")
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
