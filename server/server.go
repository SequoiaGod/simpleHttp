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

func (ms *MapStore) get() map[string]int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.peopleMap
}

func (ms *MapStore) set(key string, value int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.peopleMap[key] = value
}

func (ms *MapStore) delete(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	_, ok := ms.peopleMap[key]
	if !ok {
		return fmt.Errorf("people not found %s", key)
	}
	delete(ms.peopleMap, key)
	return nil
}

func NewServer() {
	mapStore := &MapStore{peopleMap: make(map[string]int)}
	server := &http.Server{Addr: ":8080"}
	mapStore.peopleMap["yan"] = 25
	mapStore.peopleMap["gin"] = 44
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getHandler(w, r, mapStore)
		case "POST":
			postHandler(w, r, mapStore)
		case "PUT":
			updateHandler(w, r, mapStore)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/delete/"):]
		err := mapStore.delete(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

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

func updateHandler(w http.ResponseWriter, r *http.Request, ms *MapStore) {
	people := requestDecoder(w, r)
	ms.set(people.Name, people.Age)
	fmt.Fprintln(w, people.Name, ":", people.Age)
}

func postHandler(w http.ResponseWriter, r *http.Request, ms *MapStore) {
	people := requestDecoder(w, r)
	ms.set(people.Name, people.Age)
	fmt.Fprintln(w, people.Name, ":", people.Age)
}

func getHandler(w http.ResponseWriter, r *http.Request, ms *MapStore) {
	peoples := ms.get()
	fmt.Fprintln(w, "{")
	for k, v := range peoples {
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
