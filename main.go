package main

import (
	"log"
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

// The following are TMDB urls that can be used for querying the API
const TMDB_URL string = "https://api.themoviedb.org"
const TMDB_DISCOVER string = "/3/discover/movie"


// get movies from external service and add them to the database
func populateMovieData(c *gin.Context) {
	
	// req, err := http.NewRequestWithContext(c, http.MethodGet, MOVIE_DISCOVER, nil)
	// if err != nil {
	// 	fmt.Printf("There was an error setting up the request to TMDB %s", err)
	// }
	// req.Header.Set("Authorization", "Bearer "+os.Getenv("TMDB_API_TOKEN"))
	// req.Header.Set("Accept", "application/json")
	
	// q := req.URL.Query()
	// q.Set("primary_release_year", "2025")
	// q.Set("sort_by", "primary_release_date.desc")
	// q.Set("vote_count.gte", "50")
	// q.Set("with_original_language", "en")
	// req.URL.RawQuery = q.Encode()
	
	// var client = &http.Client{
	// 	Timeout: 10 * time.Second,
	// }
	
	params := make(map[string]string)
	params["primary_release_year"] = "2025"
	params["sort_by"] = "primary_release_date.desc"
	params["vote_count.get"] = "50"
	params["with_original_language"] = "en"
	
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer "+os.Getenv(("TMDB_API_TOKEN"))
	headers["Accept"] = "application/json"
	
	client := NewHTTPClient(TMDB_URL)
	resp, body, err := client.Get(c, TMDB_DISCOVER, params, headers)
	
	
	if err != nil{
		log.Fatal(err)
	}
	
	c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
	
}

func main() {
	
	_ = godotenv.Load()
	
	router := gin.Default()
	router.GET("populate-movies", populateMovieData)
	
	router.Run("localhost:8001")
}