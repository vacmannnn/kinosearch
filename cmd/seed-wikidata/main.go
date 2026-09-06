package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const wikidataEndpoint = "https://query.wikidata.org/sparql"

type movie struct {
	ID       int
	Title    string
	Year     int
	Director string
}

type sparqlResponse struct {
	Results struct {
		Bindings []struct {
			Movie struct {
				Value string `json:"value"`
			} `json:"movie"`
			Title struct {
				Value string `json:"value"`
			} `json:"movieLabel"`
			ReleaseDate struct {
				Value string `json:"value"`
			} `json:"releaseDate"`
			Director struct {
				Value string `json:"value"`
			} `json:"director"`
		} `json:"bindings"`
	} `json:"results"`
}

func main() {
	if err := run(); err != nil {
		fmt.Println("seed failed:", err)
		os.Exit(1)
	}
}

func run() error {
	limit := flag.Int("limit", 1000, "number of movies to load")
	truncate := flag.Bool("truncate", true, "truncate movies table before insert")
	dryRun := flag.Bool("dry-run", false, "fetch movies without writing to database")
	flag.Parse()

	if *limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	movies, err := fetchMovies(ctx, *limit)
	if err != nil {
		return err
	}
	if len(movies) == 0 {
		return fmt.Errorf("wikidata returned no movies")
	}

	fmt.Printf("fetched %d movies from Wikidata\n", len(movies))
	if *dryRun {
		for _, movie := range movies[:min(len(movies), 10)] {
			fmt.Printf("%d. %s (%d), %s\n", movie.ID, movie.Title, movie.Year, movie.Director)
		}
		return nil
	}

	if err := saveMovies(ctx, movies, *truncate); err != nil {
		return err
	}

	fmt.Printf("saved %d movies to database\n", len(movies))
	return nil
}

func fetchMovies(ctx context.Context, limit int) ([]movie, error) {
	queryLimit := limit * 3
	if queryLimit < 100 {
		queryLimit = 100
	}

	form := url.Values{}
	form.Set("query", sparqlQuery(queryLimit))
	form.Set("format", "json")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wikidataEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "kinosearch-homework-seed/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikidata returned %s", resp.Status)
	}

	var data sparqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	movies := make([]movie, 0, len(data.Results.Bindings))
	seenMovies := make(map[string]struct{})
	for _, binding := range data.Results.Bindings {
		if _, ok := seenMovies[binding.Movie.Value]; ok {
			continue
		}
		seenMovies[binding.Movie.Value] = struct{}{}

		year, err := releaseYear(binding.ReleaseDate.Value)
		if err != nil {
			continue
		}

		title := strings.TrimSpace(binding.Title.Value)
		director := strings.TrimSpace(binding.Director.Value)
		if title == "" || director == "" {
			continue
		}

		movies = append(movies, movie{
			ID:       len(movies) + 1,
			Title:    title,
			Year:     year,
			Director: director,
		})
		if len(movies) == limit {
			break
		}
	}

	return movies, nil
}

func saveMovies(ctx context.Context, movies []movie, truncate bool) error {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return err
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if truncate {
		if _, err := tx.Exec(ctx, "TRUNCATE movies"); err != nil {
			return err
		}
	}

	for _, movie := range movies {
		_, err := tx.Exec(ctx, `
			INSERT INTO movies (id, title, year, director)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				year = EXCLUDED.year,
				director = EXCLUDED.director
		`, movie.ID, movie.Title, movie.Year, movie.Director)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func sparqlQuery(limit int) string {
	return fmt.Sprintf(`
SELECT ?movie ?movieLabel ?releaseDate ?director
WHERE {
  ?movie wdt:P31 wd:Q11424;
         wdt:P345 [];
         wdt:P577 ?releaseDate;
         wdt:P57 ?directorEntity;
         rdfs:label ?movieLabel.

  ?directorEntity rdfs:label ?director.

  FILTER(LANG(?movieLabel) = "en")
  FILTER(LANG(?director) = "en")
}
LIMIT %d
`, limit)
}

func releaseYear(value string) (int, error) {
	if len(value) < 4 {
		return 0, fmt.Errorf("bad release date %q", value)
	}

	return strconv.Atoi(value[:4])
}
