package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strconv"
	"time"

	"verifi-server/data"
	"verifi-server/server"

	"github.com/jung-kurt/gofpdf"
)

// RequestCollection структура запроса от клиента с номерами выполненных запросов
type RequestCollection struct {
	Links []int `json:"links_list"`
}

// reportPostHandler обрабатывает POST запрос для генерации PDF отчета
func reportPostHandler(w http.ResponseWriter, r *http.Request) {

	var req RequestCollection
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		WriterJSON(w, http.StatusBadRequest, fmt.Sprintf("невозможно прочитать тело запроса %v", err.Error()))
		return
	}

	// десериализуем запрос клиента в []int
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		WriterJSON(w, http.StatusBadRequest, fmt.Sprintf("невозможно десериализовать тело запроса %v", err.Error()))
		return
	}

	// проверяем на всякий случай
	if len(req.Links) == 0 {
		WriterJSON(w, http.StatusBadRequest, "номеров нет")
		return
	}

	// если сервер получил команду остановки/перезагрузки
	// записываем поступающие текущие запросы-номера в NumberLinksCache
	// и заканчиваем соединение
	if server.IsShutdown() {
		data.SaveNumberLinksCache(req.Links)
		WriterJSON(w, http.StatusServiceUnavailable, "сервис недоступен - повторите запрос позднее")
		return
	}
	// Проверять и дообрабатывать, если остались, номера запросов после shutdown,
	// очевидно, не имеет смысла, так как не ясно кому именно они нужны.
	// Возможно, имеет смысл добавить/уточнить логику того, что делать
	// с запросами по номерам при перезагрузке сервера.

	// собираем все данные по указанным номерам
	allResults := collectReportData(req.Links)
	if len(allResults) == 0 {
		WriterJSON(w, http.StatusNotFound, "не найдено записей по таким номерам")
		return
	}

	// генерируем PDF
	pdfData, err := generatePDF(allResults)
	if err != nil {
		WriterJSON(w, http.StatusInternalServerError, "не удалось сформировать PDF")
		return
	}

	// отправляем PDF файл
	sendPDFResponse(w, pdfData)
}

// collectReportData собирает все результаты по указанным номерам
func collectReportData(linksList []int) map[string]string {

	allResults := make(map[string]string)

	for i := range linksList {
		if results, exists := data.GetResults(linksList[i]); exists {
			maps.Copy(allResults, results)
		}
	}

	return allResults
}

// generatePDF создает PDF файл с отчетом
func generatePDF(reportData map[string]string) ([]byte, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// заголовок
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Link Status Report", "", 0, "C", false, 0, "")
	pdf.Ln(12)

	// информация о отчете
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 8, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, 8, fmt.Sprintf("Total URLs: %d", len(reportData)), "", 0, "L", false, 0, "")
	pdf.Ln(10)

	// заголовки таблицы
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(120, 10, "URL", "1", 0, "C", true, 0, "")
	pdf.CellFormat(0, 10, "Status", "1", 0, "C", true, 0, "")
	pdf.Ln(10)

	// данные
	pdf.SetFont("Arial", "", 11)
	for url, status := range reportData {
		// обрезаем слишком длинные URL для лучшего отображения
		displayURL := url
		if len(displayURL) > 50 {
			displayURL = displayURL[:47] + "..."
		}

		// URL
		pdf.CellFormat(120, 8, displayURL, "1", 0, "L", false, 0, "")

		// status с цветом
		if status == AvailableStatus {
			pdf.SetTextColor(0, 128, 0) // зеленый
			pdf.CellFormat(0, 8, "Available", "1", 0, "C", false, 0, "")
		} else {
			pdf.SetTextColor(255, 0, 0) // красный
			pdf.CellFormat(0, 8, "Not Available", "1", 0, "C", false, 0, "")
		}

		// возвращаем черный цвет для следующей строки
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(8)
	}

	// сохраняем в buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// sendPDFResponse отправляет PDF файл в ответе
func sendPDFResponse(w http.ResponseWriter, pdfData []byte) {

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}
