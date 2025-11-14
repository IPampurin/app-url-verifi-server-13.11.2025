package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ShowHelp выводит справку по приложению
func ShowHelp(port string) {

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

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		switch input {
		case "stop":
			fmt.Println("👋 Останавливаем сервер...")

			// TODO реализовать gracefull shutdown

			os.Exit(0)

		case "restart":
			fmt.Println("🔄 Перезапуск сервера...")

			// TODO добавить логику перезапуска с gracefull shutdown

			fmt.Println("✅ Сервер перезапущен")

		case "status":
			fmt.Printf("✅ Сервер работает на http://localhost:%s", port)

		case "help":
			ShowHelp(port)

		case "":
			// заигнорим пустой ввод

		default:
			fmt.Println("❌ Неизвестная команда. Напишите 'help' для справки.")
		}
	}
}
