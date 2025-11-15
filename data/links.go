package data

import (
	"sync"
)

// ShutdownCache список ссылок позапросно, переданных после команды перезагрузки или выключения
type ShutdownCache struct {
	CacheLinks [][]string
	cacheMu    sync.Mutex
}

// SDCache экземпляр ShutdownCache
var SDCache = &ShutdownCache{
	CacheLinks: make([][]string, 0),
}

// Storage структура хранилища результатов
type Storage struct {
	data   map[int]map[string]string // map [links_num] map [{url: status}]
	nextID int                       // счётчик запросов
	mu     sync.RWMutex
}

// storage экземпляр хранилища
var storage = &Storage{
	data:   make(map[int]map[string]string),
	nextID: 1,
}

// NumberLinksCache список номеров позапросно, переданных после команды перезагрузки или выключения
type NumberLinksCache struct {
	CacheNumbers [][]int
	numbersMu    sync.Mutex
}

// NLCache экземпляр NumberLinksCache
var NLCache = &NumberLinksCache{
	CacheNumbers: make([][]int, 0),
}

// SaveLinksCache сохраняет набор ссылок из запроса в ShutdownCache
func SaveLinksCache(links []string) {

	SDCache.cacheMu.Lock()
	defer SDCache.cacheMu.Unlock()

	SDCache.CacheLinks = append(SDCache.CacheLinks, links)
}

// SaveNumberLinksCache сохраняет номера запросов при shutdown
func SaveNumberLinksCache(nums []int) {

	NLCache.numbersMu.Lock()
	defer NLCache.numbersMu.Unlock()

	NLCache.CacheNumbers = append(NLCache.CacheNumbers, nums)
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
