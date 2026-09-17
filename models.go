package main

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Movie is the persisted catalog record used for discovery and recommendations.
type Movie struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	TMDBID           int            `gorm:"uniqueIndex;not null" json:"tmdb_id"`
	Title            string         `gorm:"not null" json:"title"`
	OriginalTitle    string         `json:"original_title"`
	Overview         string         `gorm:"type:text" json:"overview"`
	ReleaseDate      *time.Time     `json:"release_date"`
	PosterPath       string         `json:"poster_path"`
	BackdropPath     string         `json:"backdrop_path"`
	OriginalLanguage string         `json:"original_language"`
	VoteAverage      float32        `json:"vote_average"`
	VoteCount        int            `json:"vote_count"`
	Popularity       float32        `json:"popularity"`
	Adult            bool           `json:"adult"`
	GenreIDs         pq.Int64Array  `gorm:"type:integer[]" json:"genre_ids"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// Genre is a TMDB genre. TMDB maintains separate genre lists for movies and
// TV series that can share an id with a different meaning, or the same
// meaning under a different id, so MediaType is part of the identity.
type Genre struct {
	ID        int    `gorm:"primaryKey" json:"id"`
	MediaType string `gorm:"primaryKey" json:"media_type"` // "movie" or "tv"
	Name      string `gorm:"not null" json:"name"`
}
