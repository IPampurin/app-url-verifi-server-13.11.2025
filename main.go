package main

import (
	"fmt"

	"verifi-server/server"
)

func main() {

	var err error

	err = server.Run()
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		return
	}
}
