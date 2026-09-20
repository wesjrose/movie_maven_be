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
const TMDB_GENRE_MOVIE_LIST string = "/3/genre/movie/list"
const TMDB_GENRE_TV_LIST string = "/3/genre/tv/list"
const defaultMoviePageSize = 20
const maxMoviePageSize = 100

var tmdbGenreSources = []struct {
	path      string
	mediaType string
}{
	{TMDB_GENRE_MOVIE_LIST, "movie"},
	{TMDB_GENRE_TV_LIST, "tv"},
}

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

func upsertGenres(genres []Genre) error {
	if len(genres) == 0 {
		return nil
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}, {Name: "media_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&genres).Error
}

func populateGenres(c *gin.Context) {
	client := NewHTTPClient(TMDB_URL)
	headers := tmdbAuthHeaders()

	var genres []Genre

	for _, source := range tmdbGenreSources {
		resp, body, err := client.Get(c, source.path, nil, headers)
		if err != nil {
			log.Printf("populateGenres: failed to fetch %s genres from TMDB: %v", source.mediaType, err)
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to fetch %s genres from TMDB", source.mediaType)})
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("populateGenres: TMDB returned status %d for %s genres: %s", resp.StatusCode, source.mediaType, previewBody(body))
			c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
			return
		}

		payload, err := parseTMDBGenreList(body)
		if err != nil {
			log.Printf("populateGenres: failed to parse TMDB %s genre response: %v body=%s", source.mediaType, err, previewBody(body))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse TMDB response"})
			return
		}

		genres = append(genres, genresFromTMDB(payload.Genres, source.mediaType)...)
	}

	if err := upsertGenres(genres); err != nil {
		log.Printf("populateGenres: failed to save genres: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save genres"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saved":  len(genres),
		"genres": genres,
	})
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
		log.Printf("populateMovieData: failed to fetch movies from TMDB: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch movies from TMDB"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("populateMovieData: TMDB returned status %d: %s", resp.StatusCode, previewBody(body))
		c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
		return
	}

	payload, err := parseTMDBDiscover(body)
	if err != nil {
		log.Printf("populateMovieData: failed to parse TMDB response: %v body=%s", err, previewBody(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse TMDB response"})
		return
	}

	movies := moviesFromTMDB(payload.Results)
	if err := upsertMovies(movies); err != nil {
		log.Printf("populateMovieData: failed to save movies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saved":  len(movies),
		"movies": movies,
	})
}

func populateMoviesSince(c *gin.Context) {
	yearQuery := c.Query("year")
	year, err := strconv.Atoi(yearQuery)
	maxYear := time.Now().Year() + 1
	if err != nil || year < 1870 || year > maxYear {
		log.Printf("populateMoviesSince: invalid year query %q: %v", yearQuery, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("year query parameter must be an integer between 1870 and %d", maxYear),
		})
		return
	}

	client := NewHTTPClient(TMDB_URL)
	headers := tmdbAuthHeaders()

	sixMonthsAgo := time.Now().AddDate(0, -6, 0).Format("2006-01-02")
	log.Printf("populateMoviesSince: six months ago date is %s", sixMonthsAgo)

	params := map[string]string{
		"primary_release_date.gte": fmt.Sprintf("%d-01-01", year),
		"primary_release_date.lte": sixMonthsAgo,
		"sort_by":                  "primary_release_date.desc",
		"vote_count.gte":           "200",
		"with_original_language":   "en",
	}

	saved := 0
	pagesFetched := 0
	totalPages := 1

	for page := 1; page <= totalPages; page++ {
		params["page"] = strconv.Itoa(page)

		resp, body, err := client.Get(c, TMDB_DISCOVER, params, headers)
		if err != nil {
			log.Printf("populateMoviesSince: failed to fetch page %d for year %d: %v", page, year, err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":         "failed to fetch movies from TMDB",
				"year":          year,
				"pages_fetched": pagesFetched,
				"saved":         saved,
			})
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("populateMoviesSince: TMDB returned status %d for page %d year %d: %s", resp.StatusCode, page, year, previewBody(body))
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
			log.Printf("populateMoviesSince: failed to parse TMDB response for page %d year %d: %v body=%s", page, year, err, previewBody(body))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse TMDB response"})
			return
		}

		if page == 1 {
			log.Printf("populateMoviesSince: TMDB reported %d total pages for year %d", payload.TotalPages, year)
			totalPages = payload.TotalPages
			if totalPages > tmdbMaxDiscoverPages {
				totalPages = tmdbMaxDiscoverPages
			}
			log.Printf("populateMoviesSince: fetching %d pages for year %d", totalPages, year)
		}

		movies := moviesFromTMDB(payload.Results)
		if err := upsertMovies(movies); err != nil {
			log.Printf("populateMoviesSince: failed to save movies for page %d year %d: %v", page, year, err)
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
	page := 1
	if raw := c.Query("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			log.Printf("listMovies: invalid page query %q: %v", raw, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "page query parameter must be an integer >= 1"})
			return
		}
		page = parsed
	}

	pageSize := defaultMoviePageSize
	if raw := c.Query("page_size"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxMoviePageSize {
			log.Printf("listMovies: invalid page_size query %q: %v", raw, err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("page_size query parameter must be an integer between 1 and %d", maxMoviePageSize),
			})
			return
		}
		pageSize = parsed
	}

	var genreID int
	filterByGenre := false
	if raw := c.Query("genre_id"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			log.Printf("listMovies: invalid genre_id query %q: %v", raw, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "genre_id query parameter must be an integer"})
			return
		}
		genreID = parsed
		filterByGenre = true
	}

	countQuery := db.Model(&Movie{})
	findQuery := db.Model(&Movie{})
	if filterByGenre {
		countQuery = countQuery.Where("? = ANY(genre_ids)", genreID)
		findQuery = findQuery.Where("? = ANY(genre_ids)", genreID)
	}

	var totalResults int64
	if err := countQuery.Count(&totalResults).Error; err != nil {
		log.Printf("listMovies: failed to count movies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load movies"})
		return
	}

	movies := []Movie{}
	offset := (page - 1) * pageSize
	if err := findQuery.Order("release_date DESC NULLS LAST").Limit(pageSize).Offset(offset).Find(&movies).Error; err != nil {
		log.Printf("listMovies: failed to load movies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load movies"})
		return
	}

	totalPages := 0
	if totalResults > 0 {
		totalPages = int((totalResults + int64(pageSize) - 1) / int64(pageSize))
	}

	c.JSON(http.StatusOK, gin.H{
		"page":          page,
		"page_size":     pageSize,
		"total_pages":   totalPages,
		"total_results": totalResults,
		"movies":        movies,
	})
}

func main() {
	logFile, err := initLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	if err := godotenv.Load(); err != nil {
		log.Printf("loading .env: %v", err)
	}

	db, err = openDB()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Header("Content-Type", "application/yaml; charset=utf-8")
		c.File("openapi.yaml")
	})
	router.GET("/populate-genres", populateGenres)
	router.GET("/populate-movies", populateMovieData)
	router.GET("/populate-movies-since", populateMoviesSince)
	router.GET("/movies", listMovies)

	if err := router.Run("localhost:8001"); err != nil {
		log.Fatal(err)
	}
}
