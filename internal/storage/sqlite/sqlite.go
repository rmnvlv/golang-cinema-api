package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rmnvlv/testGolangCinema/internal/storage"
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
		name TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
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

/*
	Получение списка фильмов с сортировкой -- выбор сортировки на уровне хэндлера -- сортировка здесь
		-- подтягивать актеров по смежной таблице
	Поиск фильма по фрагменту названия, по фрагменту имени актера -- выдача списка фильмов, селект
		зависит от того что ввели на поиске -- ТЕ можно бахнуть свичкейс с тэгом а дальше селектить

	Получение списка актеров -- подтягивать фильмы по смежной таблице

	Создание правила фильм-актер: Роут создания фильмов, туда передается список актеров (какой список? Имен? Вся структура?)
		Сначала создается филь возращается id
		Создаются актеры или подтягиваются уже имеющиеся и берется id
		Каждому актеру назначается фильм в сооответсвии с id

*/

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

func (s *Storage) GetActors() ([]Actor, error) {
	// query := "SELECT * FROM actors"
	// rows, err := s.db.Query(query)
	// if err != nil {
	// 	return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.Query", err)
	// }
	// defer rows.Close()

	// var actors []Actor
	// for rows.Next() {
	// 	var actor Actor
	// 	bithString := ""
	// 	err := rows.Scan(&actor.id, &actor.name, &actor.gender, &bithString)

	// 	if err != nil {
	// 		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.RowsScan", err)
	// 	}

	// 	if bithString != "" {
	// 		date, err := time.Parse("2006-01-02", bithString[:10])

	// 		if err != nil {
	// 			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.DateConvert", err)
	// 		}
	// 		actor.birth = date
	// 	}
	// 	actors = append(actors, actor)
	// }

	// for _, v := range actors {
	// 	query, err := s.db.Prepare(`
	// 	SELECT m.name, m.id
	// 	FROM movies m
	// 	JOIN rules r ON m.id = r.movieId
	// 	JOIN actors a ON r.actorId = a.id
	// 	WHERE a.id = ?`)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.SelectMoviePrepare", err)
	// 	}

	// 	rows, err := query.Query(v.id)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.SelectMovieQuery", err)
	// 	}
	// 	defer rows.Close()
	// 	fmt.Println(rows)

	// 	var movie Movie
	// 	for rows.Next() {
	// 		err = rows.Scan(&movie.name, &movie.id)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.RowsScan", err)
	// 		}
	// 	}

	// 	v.movies = append(v.movies, movie.name)
	// }

	// return actors, nil

	query := `
	SELECT a.id, a.name, m.id, m.name
	FROM actors a
	JOIN rules r ON a.id = r.actorId
	JOIN movies m ON r.movieId = m.id
`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.Query", err)
	}
	defer rows.Close()

	actorsMap := make(map[int64]*Actor)
	for rows.Next() {
		var actorId int64
		var actorName string
		var movieId int64
		var movieName string

		err := rows.Scan(&actorId, &actorName, &movieId, &movieName)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.Scan", err)
		}

		actor, ok := actorsMap[actorId]
		if !ok {
			actor = &Actor{
				id:     actorId,
				name:   actorName,
				movies: []string{},
			}
			actorsMap[actorId] = actor
		}

		actor.movies = append(actor.movies, fmt.Sprint(Movie{id: movieId, name: movieName}))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetActors.rowsErr", err)
	}

	actors := make([]Actor, 0, len(actorsMap))
	for _, actor := range actorsMap {
		actors = append(actors, *actor)
	}

	return actors, nil
}

//Movie

func (s *Storage) CreateMovie(name string, description string, date time.Time, rate int8) (int64, error) {

	query, err := s.db.Prepare("INSERT INTO movies(name, description, date, rate) VALUES(?, ?, ?, ?)")

	if err != nil {
		return 0, fmt.Errorf("%s, %w", "storage.sqlite.CreateMovie.Prepare", err)
	}

	result, err := query.Exec(name, description, date, rate)
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

func (s *Storage) GetMoviesSorted(sortBy string) ([]Movie, error) {
	var query string

	switch sortBy {
	case "name":
		query = "SELECT * FROM movies ORDER BY name ASC"
	case "date":
		query = "SELECT * FROM movies ORDER BY date ASC"
	default:
		query = "SELECT * FROM movies ORDER BY rate DESC"
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.Query", err)
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		timeString := ""
		err := rows.Scan(&movie.id, &movie.name, &movie.description, &timeString, &movie.rate)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.RowsScan", err)
		}
		date, err := time.Parse("2006-01-02", timeString[:10])
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.DateConvert", err)
		}
		movie.date = date
		movies = append(movies, movie)
	}

	//TODO: Добавить актеров через смежную таблицу +

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.GetMovies.RowsError", err)
	}

	return movies, nil
}

func (s *Storage) GetMovieByFragment(fragmentType string, fragment map[string]interface{}) ([]Movie, error) {
	switch fragmentType {
	case "name":
		movies, err := s.searchMoviesByName(fmt.Sprint(fragment["name"]))
		return movies, err
	case "actor":
		movies, err := s.searchMoviesByActor(fmt.Sprint(fragment["actor"]))
		return movies, err
	}

	return nil, fmt.Errorf("%s", "storage.sqlite.searchMoviesByFragment.NotEnoughtFragments")
}

func (s *Storage) searchMoviesByName(fragment string) ([]Movie, error) {
	query := `
        SELECT *
        FROM movies m
        WHERE m.name LIKE ?
    `
	rows, err := s.db.Query(query, "%"+fragment+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		timeString := ""
		err := rows.Scan(&movie.id, &movie.name, &movie.description, &timeString, &movie.rate)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByName.RowsScan", err)
		}
		date, err := time.Parse("2006-01-02", timeString[:10])
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByName.DateConvert", err)
		}
		movie.date = date
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByName.RowsError", err)
	}

	return movies, nil
}

func (s *Storage) searchMoviesByActor(fragment string) ([]Movie, error) {
	query := `
        SELECT *
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

	var movies []Movie
	for rows.Next() {
		var movie Movie
		timeString := ""
		err := rows.Scan(&movie.id, &movie.name, &movie.description, &timeString, &movie.rate)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByActor.RowsScan", err)
		}
		date, err := time.Parse("2006-01-02", timeString[:10])
		if err != nil {
			return nil, fmt.Errorf("%s, %w", "storage.sqlite.searchMoviesByActor.DateConvert", err)
		}
		movie.date = date
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
		_, err = tx.Exec("INSERT INTO rules (movieId, actorId) VALUES(?, ?)", novieId, acactorId)
		if err != nil {
			return fmt.Errorf("%s, %w", "storage.sqlite.CreateRule.txFor", err)
		}
	}

	return nil
}

type Movie struct {
	id          int64
	name        string
	description string
	date        time.Time
	rate        int
	actors      []Actor
}

type Actor struct {
	id     int64
	name   string
	gender string
	birth  time.Time
	movies []string
}
