package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/rmnvlv/golang-cinema-api/internal/models"
	_ "github.com/rmnvlv/golang-cinema-api/internal/models"
	"github.com/rmnvlv/golang-cinema-api/internal/storage"
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
		3 таб
		Актеры: id, имя, пол, дата рождения
		Фильмы: id, название, описание, дата выпуска, рейтинг
		Актеры+фильмы: idMovie - idActor
	*/

	//TODO: поменять id на guid

	query, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS movies(
		id INTEGER NOT NULL PRIMARY KEY,
		title TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		date TEXT NOT NULL,
		rating INTEGER NOT NULL);
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
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
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
		movie_id INTEGER NOT NULL,
		actor_id INTEGER);
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

// Actor
func (s *Storage) CreateActor(name string, gender string, birth time.Time) (int64, error) {
	query, err := s.db.Prepare("INSERT INTO actors(name, gender, birthDate) VALUES(?, ?, ?)")

	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateActor.Prepare", err)
	}

	result, err := query.Exec(name, gender, birth)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateActor.Exec", storage.ErrFilmExists)
		}

		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateActor.Exec", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateActor.LastId", err)
	}

	return id, nil
}

func (s *Storage) UpdateActor(actorId int, updates map[string]interface{}) (int64, error) {
	queryString := "UPDATE actors SET"
	var args []interface{}

	for k, v := range updates {
		queryString += " " + k + " = ?,"
		args = append(args, v)
	}

	queryString = queryString[:len(queryString)-1]
	queryString += "WHERE id = ?"
	args = append(args, actorId)

	result, err := s.db.Exec(queryString, args...)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s, %w", "storage.sqlite.UpdateActor.Exec", storage.ErrFilmExists)
		}

		return 0, fmt.Errorf("%s, %w", "storage.sqlite.UpdateActor.Exec", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.UpdateActor.LastId", err)
	}

	return id, nil
}

func (s *Storage) DeleteActor(actorId int64) error {
	query, err := s.db.Prepare("DELETE FROM actors WHERE id = ?")

	if err != nil {
		return fmt.Errorf("%s, %w", "storage.sqlite.DeliteActor.Prepare", err)
	}

	_, err = query.Exec(actorId)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s, %w", "storage.sqlite.DeliteActor.Exec", storage.ErrFilmExists)
		}

		return fmt.Errorf("%s, %w", "storage.sqlite.DeliteActor.Exec", err)
	}

	return nil
}

func (s *Storage) GetActors() ([]models.Actor, error) {
	query := `
	SELECT a.id, a.name, m.id, m.title
	FROM actors a
	JOIN rules r ON a.id = r.actorId
	JOIN movies m ON r.movieId = m.id
`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.Query", err)
	}
	defer rows.Close()

	actorsMap := make(map[int64]*models.Actor)
	for rows.Next() {
		var actorId int64
		var actorName string
		var movieId int64
		var movieTitle string

		err := rows.Scan(&actorId, &actorName, &movieId, &movieTitle)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.Scan", err)
		}

		actor, ok := actorsMap[actorId]
		if !ok {
			actor = &models.Actor{
				Id:     actorId,
				Name:   actorName,
				Movies: []string{},
			}
			actorsMap[actorId] = actor
		}

		actor.Movies = append(actor.Movies, fmt.Sprint(models.Movie{Id: movieId, Title: movieTitle}))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.rowsErr", err)
	}

	actors := make([]models.Actor, 0, len(actorsMap))
	for _, actor := range actorsMap {
		actors = append(actors, *actor)
	}

	return actors, nil
}

//Movie

func (s *Storage) CreateMovie(title string, description string, date time.Time, rating int8) (int64, error) {

	query, err := s.db.Prepare("INSERT INTO movies(title, description, date, rating) VALUES(?, ?, ?, ?)")

	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateMovie.Prepare", err)
	}

	result, err := query.Exec(title, description, date, rating)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateMovie.Exec", storage.ErrFilmExists)
		}

		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateMovie.Exec", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateMovie.LastId", err)
	}

	return id, nil
}

func (s *Storage) UpdateMovie(filmId int, updates map[string]interface{}) (int64, error) {
	queryString := "UPDATE movies SET"
	var args []interface{}

	for k, v := range updates {
		queryString += " " + k + " = ?,"
		args = append(args, v)
	}

	queryString = queryString[:len(queryString)-1]
	queryString += "WHERE id = ?"
	args = append(args, filmId)

	result, err := s.db.Exec(queryString, args...)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s, %w", "storage.sqlite.UpdateMovie.Exec", storage.ErrFilmExists)
		}

		return 0, fmt.Errorf("%s, %w", "storage.sqlite.UpdateMovie.Exec", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.UpdateMovie.LastId", err)
	}

	return id, nil
}

func (s *Storage) DeliteMovie(filmId int) error {
	query, err := s.db.Prepare("DELETE FROM movies WHERE id = ?")

	if err != nil {
		return fmt.Errorf("%s, %w", "storage.sqlite.DeliteMovie.Prepare", err)
	}

	_, err = query.Exec(filmId)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s, %w", "storage.sqlite.DeliteMovie.Exec", storage.ErrFilmExists)
		}

		return fmt.Errorf("%s, %w", "storage.sqlite.DeliteMovie.Exec", err)
	}

	return nil
}

func (s *Storage) GetMoviesSorted(sortBy string) ([]models.Movie, error) {
	var query string
	switch sortBy {
	case "title":
		query = `
            SELECT m.id, m.title, m.rating, m.description, m.date, m.rating,  a.id, a.name, a.gender
            FROM movies m
            LEFT JOIN rules r ON m.id = r.movieId
            LEFT JOIN actors a ON r.actorId = a.id
            ORDER BY m.title ASC
        `
	case "date":
		query = `
            SELECT m.id, m.title, m.rating, m.description, m.date, m.rating,  a.id, a.name, a.gender
            FROM movies m
            LEFT JOIN rules r ON m.id = r.movieId
            LEFT JOIN actors a ON r.actorId = a.id
            ORDER BY m.date ASC
        `
	default:
		query = `
            SELECT m.id, m.title, m.description, m.date, m.rating,  a.id, a.name, a.gender
            FROM movies m
            LEFT JOIN rules r ON m.id = r.movieId
            LEFT JOIN actors a ON r.actorId = a.id
            ORDER BY m.rating DESC
        `
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.Query", err)
	}
	defer rows.Close()

	var moviesMap = make(map[int64]*models.Movie)
	for rows.Next() {
		var movieID int64
		var movieTitle string
		var movieDescription string
		var movieDateString string
		var movieRating int
		var actorID sql.NullInt64
		var actorName sql.NullString
		var actorGender string

		err := rows.Scan(&movieID, &movieTitle, &movieDescription,
			&movieDateString, &movieRating, &actorID, &actorName, &actorGender)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.Scan", err)
		}

		movieDate, err := time.Parse("2006-01-02", movieDateString[:10])
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.DateConvert", err)
		}

		movie, ok := moviesMap[movieID]
		if !ok {
			movie = &models.Movie{
				Id:          movieID,
				Title:       movieTitle,
				Description: movieDescription,
				Date:        movieDate,
				Rating:      movieRating,
				Actors:      make([]models.Actor, 0),
			}
			moviesMap[movieID] = movie
		}

		if actorID.Valid && actorName.Valid {
			movie.Actors = append(movie.Actors, models.Actor{Id: actorID.Int64, Name: actorName.String})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.RowsErr", err)
	}

	var movies = make([]models.Movie, 0, len(moviesMap))
	for _, movie := range moviesMap {
		movies = append(movies, *movie)
	}

	return movies, nil
}

func (s *Storage) GetMovieByFragment(fragmentType string, fragment string) ([]models.Movie, error) {
	switch fragmentType {
	case "title":
		movies, err := s.searchMoviesByTitle(fmt.Sprint(fragment))
		return movies, err
	case "actor":
		movies, err := s.searchMoviesByActor(fmt.Sprint(fragment))
		return movies, err
	}

	return nil, fmt.Errorf("%s", "storage.sqlite.searchMoviesByFragment.NotEnoughtFragments")
}

func (s *Storage) searchMoviesByTitle(fragment string) ([]models.Movie, error) {
	query := `
        SELECT *
        FROM movies m
        WHERE m.title LIKE ?
    `
	rows, err := s.db.Query(query, "%"+fragment+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		timeString := ""
		err := rows.Scan(&movie.Id, &movie.Title, &movie.Description, &timeString, &movie.Rating)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByTitle.RowsScan", err)
		}
		date, err := time.Parse("2006-01-02", timeString[:10])
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByTitle.DateConvert", err)
		}
		movie.Date = date
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByTitle.RowsError", err)
	}

	return movies, nil
}

func (s *Storage) searchMoviesByActor(fragment string) ([]models.Movie, error) {
	query := `
        SELECT m.id, m.title, m.description, m.date, m.rating
        FROM movies m
        JOIN rules r ON m.id = r.movieId
        JOIN actors a ON a.id = r.actorId
        WHERE a.name LIKE ?
    `
	rows, err := s.db.Query(query, "%"+fragment+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		timeString := ""
		err := rows.Scan(&movie.Id, &movie.Title, &movie.Description, &timeString, &movie.Rating)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByActor.RowsScan", err)
		}
		date, err := time.Parse("2006-01-02", timeString[:10])
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByActor.DateConvert", err)
		}
		movie.Date = date
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByActor.RowsError", err)
	}

	return movies, nil
}

//Rules

func (s *Storage) CreateRule(novieId int, actorIds []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s, %w", "storage.sqlite.CreateRule.txBegin", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	for _, acactorId := range actorIds {
		_, err = tx.Exec("INSERT INTO rules (movie_id, actor_id) VALUES(?, ?)", novieId, acactorId)
		if err != nil {
			return fmt.Errorf("%s, %w", "storage.sqlite.CreateRule.txFor", err)
		}
	}

	return nil
}
