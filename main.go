package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TMDB_URL string = "https://api.themoviedb.org"
const TMDB_DISCOVER string = "/3/discover/movie"

var db *gorm.DB

func tmdbAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + os.Getenv("TMDB_API_TOKEN"),
		"Accept":        "application/json",
	}
}

func upsertMovies(movies []Movie) error {
	if len(movies) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tmdb_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title",
			"original_title",
			"overview",
			"release_date",
			"poster_path",
			"backdrop_path",
			"original_language",
			"vote_average",
			"vote_count",
			"popularity",
			"adult",
			"genre_ids",
			"updated_at",
		}),
	}).Create(&movies).Error
}

func populateMovieData(c *gin.Context) {
	params := map[string]string{
		"primary_release_year":   "2025",
		"sort_by":                "primary_release_date.desc",
		"vote_count.gte":         "50",
		"with_original_language": "en",
	}

	client := NewHTTPClient(TMDB_URL)
	resp, body, err := client.Get(c, TMDB_DISCOVER, params, tmdbAuthHeaders())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch movies from TMDB"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
		return
	}

	payload, err := parseTMDBDiscover(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse TMDB response"})
		return
	}

	movies := moviesFromTMDB(payload.Results)
	if err := upsertMovies(movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saved":  len(movies),
		"movies": movies,
	})
}

func populateMoviesSince(c *gin.Context) {
	year, err := strconv.Atoi(c.Query("year"))
	maxYear := time.Now().Year() + 1
	if err != nil || year < 1870 || year > maxYear {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("year query parameter must be an integer between 1870 and %d", maxYear),
		})
		return
	}

	client := NewHTTPClient(TMDB_URL)
	headers := tmdbAuthHeaders()
	params := map[string]string{
		"primary_release_date.gte": fmt.Sprintf("%d-01-01", year),
		"sort_by":                  "primary_release_date.asc",
	}

	saved := 0
	pagesFetched := 0
	totalPages := 1

	for page := 1; page <= totalPages; page++ {
		params["page"] = strconv.Itoa(page)

		resp, body, err := client.Get(c, TMDB_DISCOVER, params, headers)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":         "failed to fetch movies from TMDB",
				"year":          year,
				"pages_fetched": pagesFetched,
				"saved":         saved,
			})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{
				"error":         "TMDB returned an error",
				"year":          year,
				"pages_fetched": pagesFetched,
				"saved":         saved,
			})
			return
		}

		payload, err := parseTMDBDiscover(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse TMDB response"})
			return
		}

		if page == 1 {
			totalPages = payload.TotalPages
			if totalPages > tmdbMaxDiscoverPages {
				totalPages = tmdbMaxDiscoverPages
			}
		}

		movies := moviesFromTMDB(payload.Results)
		if err := upsertMovies(movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":         "failed to save movies",
				"year":          year,
				"pages_fetched": pagesFetched,
				"saved":         saved,
			})
			return
		}

		saved += len(movies)
		pagesFetched++
	}

	c.JSON(http.StatusOK, gin.H{
		"year":          year,
		"pages_fetched": pagesFetched,
		"total_pages":   totalPages,
		"saved":         saved,
	})
}

func listMovies(c *gin.Context) {
	var movies []Movie
	if err := db.Order("release_date DESC NULLS LAST").Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

func main() {
	_ = godotenv.Load()

	var err error
	db, err = openDB()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/populate-movies", populateMovieData)
	router.GET("/populate-movies-since", populateMoviesSince)
	router.GET("/movies", listMovies)

	router.Run("localhost:8001")
}
