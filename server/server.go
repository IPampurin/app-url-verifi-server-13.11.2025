package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"net/http"
)

const timeout = 30 * time.Second // timeout для прерывания соединения при GracefulShutdown

// Server описывает HTTP сервер и его состояние
type Server struct {
	httpServer *http.Server // сам сервер
	mu         sync.RWMutex // мьютекс
	isShutdown bool         // флаг завершения работы сервера при остановке или перезапуске
}

// Srv экземпляр сервера
var Srv Server = Server{}

// Run сосредотачивет логику управления сервером
func Run(port string) error {

	// определяем сервер
	Srv.mu.Lock()
	if Srv.httpServer != nil {
		Srv.mu.Unlock()
		return fmt.Errorf("сервер уже запущен")
	}
	if Srv.isShutdown {
		Srv.mu.Unlock()
		return fmt.Errorf("сервер в процессе остановки и не может быть перезапущен")
	}
	Srv.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}
	Srv.mu.Unlock()

	// запускаем сервер в горутине
	go func() {

		fmt.Printf("сервер запущен на порту:%s\n", port)

		err := Srv.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("ошибка сервера: %v\n", err)
		}
	}()

	return nil
}

// GracefulShutdown плавно останавливает сервер
func GracefulShutdown() error {

	Srv.mu.Lock()
	if Srv.isShutdown {
		Srv.mu.Unlock()
		return nil // уже остановлен или останавливается
	}
	Srv.isShutdown = true // устанавливаем флаг
	Srv.mu.Unlock()

	// создаём контекст для принудительного обрыва соединения через разумное время
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Println("Начинаем остановку сервера...")

	if err := Srv.httpServer.Shutdown(ctx); err != nil {

		switch {
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("сервер остановлен досрочно: контекст отменён (возможно, получен сигнал ОС)")

		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("сервер не успел завершиться за %d секунд", int(timeout.Seconds()))

		default:
			// прочие ошибки (сетевые, системные)
			return fmt.Errorf("неожиданная ошибка остановки сервера: %w", err)
		}
	}

	fmt.Println("...сервер корректно остановлен.")

	return nil
}

// IsShutdown проверяет не останавливается ли сервер
func IsShutdown() bool {

	Srv.mu.RLock()
	defer Srv.mu.RUnlock()

	return Srv.isShutdown
}

// WaitForShutdownSignal ждет сигналов остановки ОС
func WaitForShutdownSignal(done <-chan struct{}) {

	// регистрируем канал отмены работы по Ctrl + C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer close(stop)

	select {
	case <-stop:
		fmt.Println("получен сигнал остановки.")

		if err := GracefulShutdown(); err != nil {
			fmt.Printf("ошибка при остановке сервера: %v\n", err)
		}
	case <-done:
		return
	}
}
