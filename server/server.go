package server

import (
	"fmt"

	"net/http"
	"os"
	"verifi-server/api"
)

func Run() error {

	port, ok := os.LookupEnv("VERIFI_PORT")
	if !ok {
		port = "8080"
	}

	api.Init()

	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
