package main

import (
	"fmt"
	"os"

	"verifi-server/cli"
	"verifi-server/server"

	"github.com/joho/godotenv"
)

func main() {

	// вычитываем env файл
	godotenv.Load(".env")

	// вычитываем номер порта из переменных, если есть
	port, ok := os.LookupEnv("VERIFI_PORT")
	if !ok {
		port = "8080"
	}

	// создаем экземпляр сервера
	srv := server.NewServer()

	// запускаем сервер в отдельной горутине
	go func() {
		err := srv.Run(port)
		if err != nil {
			fmt.Printf("Ошибка запуска сервера: %v\n", err)
			return
		}
	}()

	// показываем справку
	cli.ShowHelp(port)

	// запускаем CLI в основном потоке
	cli.RunCLI(port, srv)
}
