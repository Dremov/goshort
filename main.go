package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Link struct {
	ID          int
	OriginalUrl string
	ShortCode   string
}

func main() {
	fmt.Println("Hello World")

	var dbConnectionString = "postgres://dremov@localhost:5432/goshort"

	db, err := sql.Open("pgx", dbConnectionString)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	var originalUrl = "https://google.com"
	shortUrl, err := createShortLink(db, originalUrl)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Resulting URL is: %s\n", shortUrl)

	retrievedUrl, err := getOriginalLink(db, shortUrl)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original URL was: %s\n", retrievedUrl)
}

func createShortLink(db *sql.DB, originalUrl string) (string, error) {
	var id int
	var shortUrl = generateShortCode()
	var sqlQuery = `INSERT INTO links (original_url, short_code) 
								VALUES ($1, $2) 
								RETURNING id;`

	err := db.QueryRow(sqlQuery, originalUrl, shortUrl).Scan(&id)
	if err != nil {
		return "", err
	}

	fmt.Println("Created short link with ID: ", id)
	return shortUrl, err
}

func getOriginalLink(db *sql.DB, shortUrl string) (string, error) {
	var originalUrl string
	sqlQuery := `SELECT original_url FROM links WHERE short_code = $1;`

	err := db.QueryRow(sqlQuery, shortUrl).Scan(&originalUrl)
	if err != nil {
		return "", err
	}

	return originalUrl, nil
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
