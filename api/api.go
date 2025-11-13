package api

import "net/http"

func Init() {

	http.HandleFunc("/api/check", checkHandler)

	http.HandleFunc("/api/report", reportHandler)
}
