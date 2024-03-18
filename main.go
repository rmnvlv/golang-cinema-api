package main

import (
	"log/slog"
	"os"

	"github.com/rmnvlv/testGolangCinema/internal/config"
	"github.com/rmnvlv/testGolangCinema/internal/storage/sqlite"
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

	// Test create Movie
	// t, _ := time.Parse("2006-01-02", "2001-05-03")
	// id, err := storage.CreateMovie("Kek", "dkekdkedkekdkekd да фильмец", t, 6)
	// if err != nil {
	// 	log.Error("failed to save movie", err)
	// 	// os.Exit(1)
	// }
	// log.Debug("Movie added: %v", fmt.Sprint(id))

	// Test create actor
	// t, _ = time.Parse("2006-01-02", "1998-04-02")
	// id, err = storage.CreateActor("KOBRA", "EEEEE", t)
	// if err != nil {
	// 	log.Error("failed to save actor", err)
	// 	// os.Exit(1)
	// }
	// log.Debug("Actor added: %v", fmt.Sprint(id))

	// Test update actor
	// updateFields := map[string]interface{}{
	// 	"name":   "Penis",
	// 	"gender": "Stroooong",
	// }

	// id, err = storage.UpdateActor(4, updateFields)
	// if err != nil {
	// 	log.Error("failed to change actor", err)
	// 	os.Exit(1)
	// }
	// log.Debug("Actor changed: %v", fmt.Sprint(id))

	// idI := 2
	// err = storage.DeleteActor(int64(idI))
	// if err != nil {
	// 	log.Error("failed to delite actor", err)
	// 	os.Exit(1)
	// }
	// log.Debug("Actor deleted: %v", fmt.Sprint(idI))

	//Test update film
	// updateFields := map[string]interface{}{
	// 	"name":        "qwe",
	// 	"description": "piwo piwopiwopiwopiwopiwopiwopiwopiwopiwopiwopiwopiwopiwo",
	// }
	// id, err = storage.UpdateMovie(1, updateFields)
	// if err != nil {
	// 	log.Error("failed to update film", err)
	// 	os.Exit(1)
	// }
	// log.Debug("Film updated: %v", fmt.Sprint(id))

	//Test delite film
	// idI := 2
	// err = storage.DeliteMovie(idI)
	// if err != nil {
	// 	log.Error("failed to delite movie", err)
	// 	os.Exit(1)
	// }
	// log.Debug("Movie deleted: %v", fmt.Sprint(idI))

	//Test get movies sorted
	// movies, err := storage.GetMoviesSorted("rate")
	// if err != nil {
	// 	log.Error("failed to get movies", err)
	// 	os.Exit(1)
	// }
	// log.Debug("Movie geted: %v", movies)

	//Test get actors
	// actors, err := storage.GetActors()
	// if err != nil {
	// 	log.Error("failed to get actors", err)
	// }
	// log.Debug("Actors geted: %v", actors)

	// Test get movies by fragments
	//Name
	fragment1 := map[string]interface{}{
		"name": "1A"}
	movies1, err := storage.GetMovieByFragment("name", fragment1)
	if err != nil {
		log.Error("failed to get movies", err)
		os.Exit(1)
	}
	log.Debug("Movies by name geted: %v", movies1)
	// log.Debug("Movie geted: %v", movies)
	fragment2 := map[string]interface{}{
		"actor": "KOBRA"}
	movies2, err := storage.GetMovieByFragment("name", fragment2)
	if err != nil {
		log.Error("failed to get movies", err)
		os.Exit(1)
	}
	log.Debug("Movies by actor geted: %v", movies2)

	_ = storage

	//TODO: init router

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
