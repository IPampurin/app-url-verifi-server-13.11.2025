package server

import (
	"fmt"

	"net/http"
	"verifi-server/api"
)

func Run(port string) error {

	api.Init()

	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
