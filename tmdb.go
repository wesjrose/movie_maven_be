package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
)

const tmdbMaxDiscoverPages = 10

type tmdbDiscoverResponse struct {
	Page         int         `json:"page"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
	Results      []tmdbMovie `json:"results"`
}

type tmdbMovie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	OriginalLanguage string  `json:"original_language"`
	VoteAverage      float32 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float32 `json:"popularity"`
	Adult            bool    `json:"adult"`
	GenreIDs         []int   `json:"genre_ids"`
}

func parseTMDBDiscover(body []byte) (tmdbDiscoverResponse, error) {
	var payload tmdbDiscoverResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return tmdbDiscoverResponse{}, err
	}
	return payload, nil
}

func moviesFromTMDB(results []tmdbMovie) []Movie {
	movies := make([]Movie, 0, len(results))
	for _, result := range results {
		movies = append(movies, result.toMovie())
	}
	return movies
}

func (m tmdbMovie) toMovie() Movie {
	genreIDs := make(pq.Int64Array, len(m.GenreIDs))
	for i, id := range m.GenreIDs {
		genreIDs[i] = int64(id)
	}

	movie := Movie{
		TMDBID:           m.ID,
		Title:            m.Title,
		OriginalTitle:    m.OriginalTitle,
		Overview:         m.Overview,
		PosterPath:       m.PosterPath,
		BackdropPath:     m.BackdropPath,
		OriginalLanguage: m.OriginalLanguage,
		VoteAverage:      m.VoteAverage,
		VoteCount:        m.VoteCount,
		Popularity:       m.Popularity,
		Adult:            m.Adult,
		GenreIDs:         genreIDs,
	}

	if m.ReleaseDate != "" {
		parsed, err := time.Parse("2006-01-02", m.ReleaseDate)
		if err != nil {
			log.Printf("toMovie: invalid release_date %q for tmdb_id %d: %v", m.ReleaseDate, m.ID, err)
		} else {
			movie.ReleaseDate = &parsed
		}
	}

	return movie
}
