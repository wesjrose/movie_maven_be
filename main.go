package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TMDB_URL string = "https://api.themoviedb.org"
const TMDB_DISCOVER string = "/3/discover/movie"

var db *gorm.DB

func populateMovieData(c *gin.Context) {
	params := map[string]string{
		"primary_release_year":    "2025",
		"sort_by":                 "primary_release_date.desc",
		"vote_count.gte":          "50",
		"with_original_language":  "en",
	}

	headers := map[string]string{
		"Authorization": "Bearer " + os.Getenv("TMDB_API_TOKEN"),
		"Accept":        "application/json",
	}

	client := NewHTTPClient(TMDB_URL)
	resp, body, err := client.Get(c, TMDB_DISCOVER, params, headers)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch movies from TMDB"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
		return
	}

	results, err := parseTMDBDiscover(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse TMDB response"})
		return
	}

	movies := make([]Movie, 0, len(results))
	for _, result := range results {
		movies = append(movies, result.toMovie())
	}

	if len(movies) > 0 {
		err = db.Clauses(clause.OnConflict{
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
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save movies"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"saved":  len(movies),
		"movies": movies,
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
	router.GET("/movies", listMovies)

	router.Run("localhost:8001")
}
