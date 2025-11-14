package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"net/http"
	"verifi-server/api"
)

// Server описывает HTTP сервер и его состояние
type Server struct {
	httpServer *http.Server
	mu         sync.RWMutex
	isRunning  bool
}

// NewServer создает новый экземпляр сервера
func NewServer() *Server {

	return &Server{
		isRunning: false,
	}
}

// Run сосредотачивет логику управления сервером
func (s *Server) Run(port string) error {

	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("сервер уже работает")
	}

	api.Init()

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: nil, // используем DefaultServeMux из http.HandleFunc
	}

	s.isRunning = true
	s.mu.Unlock()

	// запускаем сервер в горутине
	go func() {
		fmt.Printf("сервер запущен на порту:%s\n", port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ошибка сервера: %v\n", err)

			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()
		}
	}()

	// ждем сигналов остановки
	s.waitForShutdownSignal()

	return nil
}

// GracefulShutdown плавно останавливает сервер
func (s *Server) GracefulShutdown() error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning || s.httpServer == nil {
		return fmt.Errorf("сервер не работает")
	}

	fmt.Println("начинаем остановку сервера...")

	// даем 10 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("ошибка остановки сервера: %v", err)
	}

	s.isRunning = false

	fmt.Println("...сервер корректно остановлен")

	return nil
}

// IsRunning проверяет запущен ли сервер
func (s *Server) IsRunning() bool {

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isRunning
}

// waitForShutdownSignal ждет сигналов остановки
func (s *Server) waitForShutdownSignal() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	fmt.Println("получен сигнал остановки.")

	if err := s.GracefulShutdown(); err != nil {
		fmt.Printf("ошибка при остановке сервера: %v\n", err)
	}
}
