package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scriptdealer/to-do-go/internal/services"
	"github.com/scriptdealer/to-do-go/internal/storage"
	"github.com/scriptdealer/to-do-go/internal/transport/rest"
)

type Configuration struct {
	ServerIP   string
	ServerPort string
}

func getConfig() *Configuration {
	cfg := Configuration{
		ServerIP:   "0.0.0.0",
		ServerPort: "8080",
	}
	ip, found := os.LookupEnv("TODO_IP")
	if found {
		cfg.ServerIP = ip
	}
	port, found := os.LookupEnv("TODO_PORT")
	if found {
		cfg.ServerPort = port
	}
	return &cfg
}

func main() {
	//InitConfig
	config := getConfig()

	//InitLogger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	//Injections
	// db := storage.NewMemoryStorage()
	db2, err := storage.NewPostgresStore(logger)
	if err != nil {
		os.Exit(1)
	}
	services := services.NewComposite(db2, logger)
	gorillaMux := rest.InitHandlers(services)
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", config.ServerIP, config.ServerPort),
		Handler:        gorillaMux,
		ReadTimeout:    14 * time.Second,
		WriteTimeout:   14 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	logger.Info("Info: Starting web/http server...", slog.String("address", server.Addr))
	// setup signal catching
	signal.Notify(services.Interruption, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-services.Interruption
		logger.Info("Interrupted", slog.Any("signal", s))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//shutdown the server
		err := server.Shutdown(ctx)
		if err == nil {
			os.Exit(0)
		} else {
			logger.Info("Graceful shutdown", slog.String("error", err.Error()))
			server.Close()
		}
	}()
	servingError := server.ListenAndServe()
	logger.Info("Exiting", slog.String("reason", servingError.Error()))
}
