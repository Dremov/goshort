package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Link struct {
	ID          int
	OriginalUrl string
	ShortCode   string
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortUrl string `json:"short_code"`
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

	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		// A. Parse the Short Code
		// r.URL.Path returns the path like "/abc12"
		// We slice it [1:] to remove the leading slash, leaving "abc12"
		shortPath := request.URL.Path[1:]

		// If the code is empty (user visited just "/"), just print a message
		if shortPath == "" {
			fmt.Fprintf(responseWriter, "Welcome to GoShort! Use the API to create links.")
			return
		}

		longPath, err := getOriginalLink(db, shortPath)
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusNotFound)
			return
		}

		// D. Redirect the User
		//    If found, use this function to send them to the original URL:
		//    http.Redirect(w, r, originalUrl, http.StatusFound)
		http.Redirect(responseWriter, request, longPath, http.StatusFound)
	})

	http.HandleFunc("/shorten", func(responseWriter http.ResponseWriter, request *http.Request) {
		// Paste the logic from step 2 here so it can see 'db'
		handleShorten(responseWriter, request, db)
	})

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

func handleShorten(responseWriter http.ResponseWriter, request *http.Request, db *sql.DB) {
	// 1. Validate Method
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Parse JSON Body
	var shortenRequest ShortenRequest
	// We create a decoder that reads from the Request Body and decodes into our struct
	if err := json.NewDecoder(request.Body).Decode(&shortenRequest); err != nil {
		http.Error(responseWriter, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// --- YOUR CODE HERE ---

	// 3. Call your createShortLink function using req.URL
	//    (Handle the error if it fails)
	shortUrl, err := createShortLink(db, shortenRequest.URL)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Send JSON Response
	//    A. Create an instance of ShortenResponse with the new code
	//    B. Set Header: w.Header().Set("Content-Type", "application/json")
	//    C. Encode it: json.NewEncoder(w).Encode(response)
	shortenResponse := ShortenResponse{
		ShortUrl: shortUrl,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(shortenResponse)
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
