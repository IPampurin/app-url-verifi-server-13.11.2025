package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const timeout = 30 * time.Second // timeout для прерывания соединения при GracefulShutdown

// Server описывает HTTP сервер и его состояние
type Server struct {
	httpServer *http.Server // сам сервер
	Mu         sync.RWMutex // мьютекс
	IsShutdown bool         // флаг завершения работы сервера при остановке или перезапуске
}

// Srv экземпляр сервера
var Srv Server = Server{}

// Run сосредотачивет логику управления сервером
func Run(port string) error {

	// определяем сервер
	Srv.Mu.Lock()
	if Srv.httpServer != nil {
		Srv.Mu.Unlock()
		return fmt.Errorf("сервер уже запущен")
	}
	if Srv.IsShutdown {
		Srv.Mu.Unlock()
		return fmt.Errorf("сервер в процессе остановки и пока не может быть перезапущен")
	}
	Srv.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}
	Srv.Mu.Unlock()

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

	Srv.Mu.Lock()
	if Srv.IsShutdown {
		Srv.Mu.Unlock()
		return nil // уже остановлен или останавливается
	}
	Srv.IsShutdown = true // устанавливаем флаг
	Srv.Mu.Unlock()

	// создаём контекст для принудительного обрыва соединения через разумное время
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Println("Начинаем остановку сервера...")

	// Shutdown "плавно" разрывает соединение по факту отсутствия обращения или через timeout
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

	Srv.Mu.RLock()
	defer Srv.Mu.RUnlock()

	return Srv.IsShutdown
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
