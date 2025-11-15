package data

import (
	"sync"
)

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

// ShutdownCache список ссылок позапросно, переданных после команды перезагрузки или выключения
type ShutdownCache struct {
	cacheLinks [][]string
	cacheMu    sync.Mutex
}

// заведем экземпляр ShutdownCache
var shutdownCache = &ShutdownCache{
	cacheLinks: make([][]string, 0),
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

func SaveLinksCache(links []string) {

	shutdownCache.cacheMu.Lock()
	defer shutdownCache.cacheMu.Unlock()

	shutdownCache.cacheLinks = append(shutdownCache.cacheLinks, links)
}

// ReadLinksCache возвращает соответствующий набор ссылок из поступивших после начала Shutdown
func ReadLinksCache(i int) []string {

	return shutdownCache.cacheLinks[i]
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

// GetResults смотрит, что есть в хранилище по номеру
func GetResults(id int) (map[string]string, bool) {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	results, exists := storage.data[id]

	return results, exists
}
