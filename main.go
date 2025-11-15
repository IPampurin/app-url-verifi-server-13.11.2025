package main

import (
	"fmt"
	"os"

	"verifi-server/api"
	"verifi-server/cli"
	"verifi-server/server"

	"github.com/joho/godotenv"
)

func main() {

	// вычитываем env файл, если есть
	godotenv.Load(".env")

	// вычитываем номер порта из переменных, если есть
	port, ok := os.LookupEnv("VERIFI_PORT")
	if !ok {
		port = "8080"
	}

	// запускаем api
	api.Init()

	// запускаем сервер
	err := server.Run(port)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		return
	}

	// запускаем CLI
	cli.RunCLI(port)
}
