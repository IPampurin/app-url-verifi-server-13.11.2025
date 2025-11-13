package api

import "net/http"

// checkHandler распределяет запросы эндпойнта "/api/check" по типу
// в данном случае у нас только POST
func checkHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		checPostkHandler(w, r)

	default:
		// WriterJSON(w, http.StatusMethodNotAllowed, "error")
	}

}
