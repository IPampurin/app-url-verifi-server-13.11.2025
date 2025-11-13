package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"verifi-server/data"
)

// checkPostHandler принимает запрос с адресами и синхронно собирает статусы
func checkPostHandler(w http.ResponseWriter, r *http.Request) {

	var req data.RequestLinks
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

	// проверяем доступность каждой ссылки
	statusLinks := make(map[string]string)
	for _, url := range req.Links {
		if isAvailable(url) {
			statusLinks[url] = data.AvailableStatus
		} else {
			statusLinks[url] = data.NotAvailableStatus
		}
	}

	// сохраняем результаты и получаем номер
	linksSetNum := data.SaveResults(statusLinks)

	// формируем и возвращаем ответ
	resp := data.ResponseLinks{
		Links:    statusLinks,
		LinksNum: linksSetNum,
	}

	WriterJSON(w, http.StatusOK, resp)
}

// isAvailable проверяет доступность URL
func isAvailable(url string) bool {

	normalizedURL := normalizeURL(url)

	// поднимаем клиента с паузой 3 секунды
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	// отправляем запрос и читаем ответ
	resp, err := client.Get(normalizedURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// считаем статусы 2xx и 3xx доступными
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// normalizeURL добавляет http:// если отсутствует
func normalizeURL(url string) string {

	if strings.Contains(url, "://") {
		return url
	}
	return "http://" + url
}
