package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"verifi-server/api"
	"verifi-server/server"
)

// Mock-сервер для имитации внешних URL
func startMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestCheckPostHandler(t *testing.T) {
	// Запускаем mock-сервер
	mock := startMockServer()
	defer mock.Close()

	// Подменяем базовый URL для тестов
	baseURL := strings.TrimSuffix(mock.URL, "/")

	tests := []struct {
		name           string
		requestBody    string
		serverShutdown bool
		expectedStatus int
		expectedLinks  map[string]string
	}{
		{
			name:           "корректный запрос с доступными/недоступными ссылками",
			requestBody:    `{"links": ["` + baseURL + `/ok", "` + baseURL + `/bad"]}`,
			serverShutdown: false,
			expectedStatus: http.StatusOK,
			expectedLinks: map[string]string{
				baseURL + "/ok":  api.AvailableStatus,
				baseURL + "/bad": api.NotAvailableStatus,
			},
		},
		{
			name:           "пустой список ссылок",
			requestBody:    `{"links": []}`,
			serverShutdown: false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "сервер в процессе остановки",
			requestBody:    `{"links": ["` + baseURL + `/ok"]}`,
			serverShutdown: true,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "некорректный JSON",
			requestBody:    `{invalid json}`,
			serverShutdown: false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Имитируем состояние сервера
			if tt.serverShutdown {
				server.Srv.Mu.Lock()
				server.Srv.IsShutdown = true
				server.Srv.Mu.Unlock()
			} else {
				server.Srv.Mu.Lock()
				server.Srv.IsShutdown = false
				server.Srv.Mu.Unlock()
			}

			// Создаём запрос
			req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBufferString(tt.requestBody))
			rec := httptest.NewRecorder()

			// Вызываем обработчик
			api.CheckPostHandler(rec, req)

			// Проверяем статус
			if rec.Code != tt.expectedStatus {
				t.Errorf("ожидали статус %d, получили %d", tt.expectedStatus, rec.Code)
			}

			// Для успешных ответов проверяем тело
			if tt.expectedStatus == http.StatusOK {
				var resp api.ResponseLinks
				err := json.NewDecoder(rec.Body).Decode(&resp)
				if err != nil {
					t.Fatal("не удалось декодировать ответ:", err)
				}

				if len(resp.Links) != len(tt.expectedLinks) {
					t.Errorf("ожидали %d ссылок, получили %d", len(tt.expectedLinks), len(resp.Links))
				}

				for url, status := range tt.expectedLinks {
					if resp.Links[url] != status {
						t.Errorf("для %s ожидали статус %s, получили %s", url, status, resp.Links[url])
					}
				}
			}
		})
	}
}

func TestIsAvailable(t *testing.T) {
	mock := startMockServer()
	defer mock.Close()

	baseURL := strings.TrimSuffix(mock.URL, "/")

	tests := []struct {
		url      string
		expected bool
	}{
		{baseURL + "/ok", true},
		{baseURL + "/bad", false},
		{"invalid-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := api.IsAvailable(tt.url)
			if result != tt.expected {
				t.Errorf("IsAvailable(%s) = %v, ожидаем %v", tt.url, result, tt.expected)
			}
		})
	}
}
