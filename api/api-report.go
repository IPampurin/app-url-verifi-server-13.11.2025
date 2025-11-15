package api

import "net/http"

// reportHandler распределяет запросы эндпойнта "/api/report" по типу
// в данном случае у нас только POST
func reportHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		ReportPostHandler(w, r)

	default:
		WriterJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
