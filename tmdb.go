package main

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type tmdbDiscoverResponse struct {
	Results []tmdbMovie `json:"results"`
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

func parseTMDBDiscover(body []byte) ([]tmdbMovie, error) {
	var payload tmdbDiscoverResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload.Results, nil
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
		if parsed, err := time.Parse("2006-01-02", m.ReleaseDate); err == nil {
			movie.ReleaseDate = &parsed
		}
	}

	return movie
}
