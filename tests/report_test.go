package tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"verifi-server/api"
)

func TestReportPostHandler_EmptyAndFakeIDs(t *testing.T) {
	testCases := []struct {
		name         string
		requestBody  string
		expectStatus int
		expectInBody string
	}{
		{
			name:         "пустой список номеров",
			requestBody:  `{"links_list": []}`,
			expectStatus: http.StatusBadRequest,
			expectInBody: "номеров нет",
		},
		{
			name:         "список с несуществующими номерами",
			requestBody:  `{"links_list": [13, 66]}`,
			expectStatus: http.StatusNotFound,
			expectInBody: "не найдено записей по таким номерам",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Создаём HTTP-запрос
			req := httptest.NewRequest(
				http.MethodPost,
				"/report",
				bytes.NewBufferString(tc.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")

			// 2. Захватываем ответ
			rec := httptest.NewRecorder()
			api.ReportPostHandler(rec, req)

			// 3. Проверяем статус-код
			if rec.Code != tc.expectStatus {
				t.Errorf("ожидали статус %d, получили %d", tc.expectStatus, rec.Code)
			}

			// 4. Читаем тело ответа
			bodyBytes, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Fatal("не удалось прочитать тело ответа:", err)
			}
			bodyStr := string(bodyBytes)

			// 5. Проверяем, что в теле есть ожидаемая подстрока
			if !strings.Contains(bodyStr, tc.expectInBody) {
				t.Errorf("в ответе не найдено: %q\nвесь ответ: %s", tc.expectInBody, bodyStr)
			}

			// 6. Дополнительно проверяем Content-Type (должен быть application/json для ошибок)
			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("ожидали Content-Type 'application/json', получили %q", contentType)
			}
		})
	}
}
