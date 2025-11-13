package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"verifi-server/data"
)

// checPostHandler принимает запрос с адресами и синхронно собирает статусы
func checkPostHandler(w http.ResponseWriter, r *http.Request) {

	var req data.Request
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		WriterJSON(w, http.StatusBadRequest, fmt.Sprintf("невозможно прочитать тело запроса %v", err.Error()))
		return
	}

	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		WriterJSON(w, http.StatusBadRequest, fmt.Sprintf("невозможно десериализовать тело запроса %v", err.Error()))
		return
	}

	resp := make([]data.Response, len(req.Links), len(req.Links))

	for i := range req.Links {

	}
}
