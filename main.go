package main

import (
	"fmt"
	"net/http"
	"os"

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

// get movies from external service and add them to the database
func populateMovieData(c *gin.Context) {
	_ = godotenv.Load()
	tmdbToken := os.Getenv("TMDB_API_TOKEN")
	fmt.Printf("The movie database token is: %s", tmdbToken)
	c.JSON(http.StatusOK, gin.H{"message" : "hello world"})
	
}

func main() {
	
	// _ = godotenv.Load()
	// tmdbToken = os.Getenv("TMDB_API_TOKEN")
	
	router := gin.Default()
	router.GET("populate-movies", populateMovieData)
	
	router.Run("localhost:8000")
}