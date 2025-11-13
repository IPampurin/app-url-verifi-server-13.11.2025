package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strconv"

	"verifi-server/data"
)

// reportPostHandler обрабатывает POST запрос для генерации PDF отчета
func reportPostHandler(w http.ResponseWriter, r *http.Request) {

	var req data.RequestCollection
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

	var content bytes.Buffer
	content.WriteString("Link Status Report\n")
	content.WriteString("==================\n\n")

	for url, status := range reportData {
		content.WriteString(fmt.Sprintf("%s: %s\n", url, status))
	}

	return content.Bytes(), nil
}

// sendPDFResponse отправляет PDF файл в ответе
func sendPDFResponse(w http.ResponseWriter, pdfData []byte) {

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}
