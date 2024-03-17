package main

import (
	"log/slog"
	"os"

	"github.com/rmnvlv/testGolangCinema/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	//init config: cleanenv
	cfg := config.MustLoad()
	// log.Print(cfg)

	//init logger: slog
	log := initLogger(cfg.Env)
	log.Info("Loger init completed", slog.String("env", cfg.Env))

	//TODO: init storage: psql

	//TODO: run server

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
