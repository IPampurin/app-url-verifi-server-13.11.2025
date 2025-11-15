package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"verifi-server/data"
	"verifi-server/server"
)

const (
	// время на запрос клиенту
	timeoutClient = 3
	// константы статусов на выдачу
	AvailableStatus    = "available"
	NotAvailableStatus = "not available"
)

// Link описывает структуру обрабатывемой ссылки
type Link struct {
	Url    string // адрес
	Status string // статус ресурса по адресу
}

// RequestLinks структура запроса от клиента со ссылками
type RequestLinks struct {
	Links []string `json:"links"`
}

// ResponseLinks структура ответа по запросу со ссылками
type ResponseLinks struct {
	Links    map[string]string `json:"links"`     // map [{url: status}]
	LinksNum int               `json:"links_num"` // номер набора
}

// checkPostHandler принимает запрос с адресами и синхронно собирает статусы
func checkPostHandler(w http.ResponseWriter, r *http.Request) {

	var req RequestLinks
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		WriterJSON(w, http.StatusBadRequest, fmt.Sprintf("невозможно прочитать тело запроса %v", err.Error()))
		return
	}

	// десериализуем запрос клиента в []string
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		WriterJSON(w, http.StatusBadRequest, fmt.Sprintf("невозможно десериализовать тело запроса %v", err.Error()))
		return
	}

	// проверяем на всякий случай
	if len(req.Links) == 0 {
		WriterJSON(w, http.StatusBadRequest, "ссылок нет")
		return
	}

	// если сервер получил команду остановки/перезагрузки
	// записываем поступающие текущие запросы-ссылки в ShutdownCache
	// и заканчиваем соединение
	if server.IsShutdown() {
		data.SaveLinksCache(req.Links)
		WriterJSON(w, http.StatusServiceUnavailable, "сервис недоступен - повторите запрос позднее")
		return
	}

	// если сервер не останавливают/перезагружают
	// проверяем доступность каждой ссылки
	statusLinks, linksSetNum := currentLinksCheck(req.Links)

	// формируем и возвращаем ответ
	resp := ResponseLinks{
		Links:    statusLinks,
		LinksNum: linksSetNum,
	}

	WriterJSON(w, http.StatusOK, resp)
}

// currentLinksCheck асинхронно проверяет доступность по текущему набору ссылок
func currentLinksCheck(links []string) (map[string]string, int) {

	statusLinks := make(map[string]string)
	results := make(chan Link, len(links))

	// запускаем проверки
	for _, url := range links {
		go func(u string) {
			if IsAvailable(u) {
				results <- Link{u, AvailableStatus}
			} else {
				results <- Link{u, NotAvailableStatus}
			}
		}(url)
	}

	// собираем результаты
	for i := 0; i < len(links); i++ {
		res := <-results
		statusLinks[res.Url] = res.Status
	}

	// сохраняем результаты и получаем номер
	linksSetNum := data.SaveResults(statusLinks)

	return statusLinks, linksSetNum
}

// IsAvailable проверяет доступность URL
func IsAvailable(url string) bool {

	// добавляем http:// если отсутствует
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}

	// поднимаем клиента со временем ожидания timeoutClient секунд
	client := http.Client{
		Timeout: timeoutClient * time.Second,
	}

	// отправляем запрос и читаем ответ
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// считаем статусы 2xx и 3xx доступными
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// CacheLinksCheck асинхронно проверяет доступность по набору ссылок
func CacheLinksCheck(links []string) {

	statusLinks := make(map[string]string)
	results := make(chan Link, len(links))

	// запускаем проверки
	for _, url := range links {
		go func(u string) {
			if IsAvailable(u) {
				results <- Link{u, AvailableStatus}
			} else {
				results <- Link{u, NotAvailableStatus}
			}
		}(url)
	}

	// собираем результаты
	for i := 0; i < len(links); i++ {
		res := <-results
		statusLinks[res.Url] = res.Status
	}

	// сохраняем результаты и игнорируем номер
	_ = data.SaveResults(statusLinks)
}
