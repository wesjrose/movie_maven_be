package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Movie struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Genre string `json:"genre"`
	Year int `json:"year"`
	Rating float32 `json:"rating"`
}

// The following are TMDB urls that can be used for querying the API
const MOVIE_DISCOVER string = "https://api.themoviedb.org/3/discover/movie"


// get movies from external service and add them to the database
func populateMovieData(c *gin.Context) {
	
	req, err := http.NewRequestWithContext(c, http.MethodGet, MOVIE_DISCOVER, nil)
	if err != nil {
		fmt.Printf("There was an error setting up the request to TMDB %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("TMDB_API_TOKEN"))
	req.Header.Set("Accept", "application/json")
	
	q := req.URL.Query()
	q.Set("primary_release_year", "2025")
	q.Set("sort_by", "primary_release_date.desc")
	q.Set("vote_count.gte", "100")
	q.Set("with_original_language", "en")
	req.URL.RawQuery = q.Encode()
	
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	
	c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
	
}

func main() {
	
	_ = godotenv.Load()
	
	router := gin.Default()
	router.GET("populate-movies", populateMovieData)
	
	router.Run("localhost:8000")
}