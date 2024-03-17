package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Path", err)
	}

	/*
		3 таблцы
		Актеры: id, имя, пол, дата рождения
		Фильмы: id, название, описание, дата выпуска, рейтинг
		Актеры+фильмы: idMovie - idActor
	*/
	query, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS movies(
		id INTIGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		discription TEXT NOT NULL,
		date TEXT NOT NULL,
		rate INTEGER NOT NULL);
		`)

	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Prepare.Movies", err)
	}

	_, err = query.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Exec.Movies", err)
	}

	query, err = db.Prepare(`
	CREATE TABLE IF NOT EXISTS actors(
		id iINTEGER PRIMARY KEY,
		name TEXT,
		gender TEXT,
		birthDate TEXT);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Prepare.Actors", err)
	}

	_, err = query.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Exec.Actors", err)
	}

	query, err = db.Prepare(`
	CREATE TABLE IF NOT EXISTS rules(
		movieId INTEGER NOT NULL,
		actorId INTEGER);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Prepare.Rules", err)
	}

	_, err = query.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.New.Exec.Rules", err)
	}

	return &Storage{db: db}, nil
}
