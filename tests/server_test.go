package tests

import (
	"net/http"
	"testing"
	"time"

	"verifi-server/server"
)

func TestServerStartupAndShutdown(t *testing.T) {
	const testPort = "8081"
	const address = "http://localhost:" + testPort

	// Запуск сервера
	if err := server.Run(testPort); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond) // Даём серверу стартовать

	// Проверка: сервер отвечает
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(address)
	if err != nil {
		t.Fatal("server should be running:", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want status 404; got %d", resp.StatusCode)
	}

	// Остановка сервера
	if err := server.GracefulShutdown(); err != nil {
		t.Fatal("GracefulShutdown failed:", err)
	}

	time.Sleep(100 * time.Millisecond) // Ждём завершения

	// Проверка: сервер больше не отвечает
	resp, err = client.Get(address)
	if err == nil {
		resp.Body.Close()
		t.Fatal("server should be stopped, but still responding")
	}
}
