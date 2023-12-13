package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type IMDB struct {
	items map[string]string
	mu    sync.RWMutex
}

func newDB() IMDB {
	file, err := os.Open("db.json")
	if err != nil {
		return IMDB{items: map[string]string{}}
	}
	items := map[string]string{}
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		fmt.Println("Could not decode the DB backup", err.Error())
		return IMDB{items: items}
	}
	return IMDB{items: items}
}

func (m IMDB) set(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = value
}

func (m IMDB) get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, found := m.items[key]
	return value, found
}

func (m IMDB) delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

func (m IMDB) save() {
	file, err := os.Create("db.json")
	if err != nil {
		fmt.Println("Could not create file ", err.Error())
	}
	if err := json.NewEncoder(file).Encode(m.items); err != nil {
		fmt.Println("Could not encode DB data for the backup ", err.Error())
	}
	fmt.Println("Successfully saved DB to a file")
}