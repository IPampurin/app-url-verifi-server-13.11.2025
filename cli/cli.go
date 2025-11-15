package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"verifi-server/api"
	"verifi-server/data"
	"verifi-server/server"
)

// showHelp выводит справку по приложению
func showHelp(port string) {

	fmt.Println("🚀 Link Verifier Server запущен!")
	fmt.Printf("🌐 Сервер доступен по адресу: http://localhost:%s\n", port)
	fmt.Println("")
	fmt.Println("Доступные команды:")
	fmt.Println("  stop     - Остановить сервер и выйти")
	fmt.Println("  restart  - Перезапустить сервер")
	fmt.Println("  status   - Показать статус сервера")
	fmt.Println("  help     - Показать эту справку")
	fmt.Println("")
	fmt.Println("Эндпоинты API:")
	fmt.Println("  POST /api/check    - Проверить доступность ссылок")
	fmt.Println("  POST /api/report   - Сгенерировать PDF отчет")
	fmt.Println("")
}

// RunCLI позволяет управлять приложением из консоли
func RunCLI(port string) {

	// канал для остановки WaitForShutdownSignal при остановке сервера
	done := make(chan struct{})

	// запускаем контроль сигналов ОС
	go server.WaitForShutdownSignal(done)

	showHelp(port)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		switch input {

		case "stop": // Graceful shutdown серверу

			if err := server.GracefulShutdown(); err != nil {
				fmt.Printf("Ошибка остановки: %v\n", err)
			} else {
				close(done)
				fmt.Println("👋 Выходим из программы.")
			}

			os.Exit(0)

		case "restart": // Graceful shutdown серверу и новый запуск

			fmt.Println("🔄 Перезапуск сервера...")

			if err := server.GracefulShutdown(); err != nil {
				fmt.Printf("Ошибка остановки: %v\n", err)

			} else {
				fmt.Println("👋 Запускаем сервер.")

				err := server.Run(port)
				if err != nil {
					fmt.Printf("Ошибка запуска сервера: %v\n", err)
					return
				}

				// проверяем и дообрабатываем, если остались, ссылки после shutdown
				if len(data.SDCache.CacheLinks) != 0 {
					for i := range data.SDCache.CacheLinks {
						api.CacheLinksCheck(data.SDCache.CacheLinks[i])
					}
					data.SDCache.CacheLinks = make([][]string, 0)
				}

				// Проверять и дообрабатывать, если остались, номера запросов после shutdown,
				// очевидно, не имеет смысла, так как не ясно кому именно они нужны.
				// Возможно, имеет смысл добавить/уточнить логику того, что делать
				// с запросами по номерам при перезагрузке сервера.
			}

			fmt.Println("✅ Сервер перезапущен")

		case "status":

			fmt.Printf("✅ Сервер работает на http://localhost:%s", port)

		case "help":

			showHelp(port)

		case "":

			// заигнорим пустой ввод

		default:
			fmt.Println("❌ Неизвестная команда. Напишите 'help' для справки.")
		}
	}
}
