package data

import "sync"

// константы статусов на выдачу
const (
	AvailableStatus    = "available"
	NotAvailableStatus = "not available"
)

// RequestLinks структура запроса от клиента со ссылками
type RequestLinks struct {
	Links []string `json:"links"`
}

// ResponseLinks структура ответа по запросу со ссылками
type ResponseLinks struct {
	Links    map[string]string `json:"links"`     // map [{url: status}]
	LinksNum int               `json:"links_num"` // номер набора
}

// RequestCollection структура запроса от клиента с номерами выполненных запросов
type RequestCollection struct {
	Links []int `json:"links_list"`
}

// Storage структура хранилища результатов
type Storage struct {
	data   map[int]map[string]string // map [links_num] map [{url: status}]
	nextID int                       // счётчик запросов
	mu     sync.RWMutex
}

// заведём экземпляр хранилища
var storage = &Storage{
	data:   make(map[int]map[string]string),
	nextID: 1,
}

// SaveResults сохраняет результаты запросов по номерам
func SaveResults(results map[string]string) int {

	storage.mu.Lock()
	defer storage.mu.Unlock()

	id := storage.nextID
	storage.data[id] = results
	storage.nextID++

	return id
}

// GetResults смотрит, что есть в хранилище
func GetResults(id int) (map[string]string, bool) {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	results, exists := storage.data[id]

	return results, exists
}
