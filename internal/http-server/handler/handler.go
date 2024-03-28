package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rmnvlv/golang-cinema-api/internal/models"
)

type MovieGetter interface {
	GetMovieByFragment(fragmentType string, fragment string) ([]models.Movie, error)
	GetMoviesSorted(sortBy string) ([]models.Movie, error)
}

func New(log *slog.Logger, s MovieGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var params models.SerchMovieParams
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			log.Error("handler.New.MovieGetter.JsonUmmarshal", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("request body decoded", slog.Any("request-body", params))

		var movies []models.Movie

		switch params.Sort {
		case true:
			movies, err = s.GetMoviesSorted(params.SortType)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("handler.New.MovieGetter.ParamsSortTrue", err)
				return
			}
		case false:
			movies, err = s.GetMovieByFragment(params.FragmentType, params.Fragments)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("handler.New.MovieGetter.ParamsSortFalse", err)
				return
			}
		default:
			movies, err = s.GetMoviesSorted("")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("handler.New.MovieGetter.Default", err)
				return
			}
		}

		jsonResp, err := json.Marshal(movies)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("handler.New.MovieGetter.MarshalJson", err)
			return
		}

		w.Write(jsonResp)
	}
}
