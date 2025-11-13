package server

import (
	"fmt"

	"net/http"
	"os"
)

func Run() error {

	port, ok := os.LookupEnv("VERIFI_PORT")
	if !ok {
		port = "8080"
	}

	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
