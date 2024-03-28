package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/rmnvlv/golang-cinema-api/internal/config"
	"github.com/rmnvlv/golang-cinema-api/internal/http-server/handler"
	_ "github.com/rmnvlv/golang-cinema-api/internal/http-server/logger"
	"github.com/rmnvlv/golang-cinema-api/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	//init config: cleanenv
	cfg := config.MustLoad()

	//init logger: slog
	log := initLogger(cfg.Env)
	log.Info("Loger init completed", slog.String("env", cfg.Env))

	//init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed with init storage", err)
		os.Exit(1)
	}
	log.Info("Storage init complited", slog.String("storage", cfg.StoragePath))

	_ = storage

	//init router
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.New(log, storage))

	log.Debug("Initializing server...", slog.String("Server: ", cfg.Address))
	//run server
	http.ListenAndServe(cfg.Address, mux)
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
