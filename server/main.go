package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

var db *sql.DB

func main() {
	initDatabase()
	// Close the database connection upon exit
	defer db.Close() 

	fmt.Println("Server listening...")
	http.HandleFunc("/cotacao", getCotacaoHandler)
	http.ListenAndServe(":8080", nil)

}

func initDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalf("Failed to open the database: %v", err)
	}

	// Ensure the database connection is healthy
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Create the table if it doesn't exist
	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY,
		bid TEXT
	)`)
	if err != nil {
		log.Fatalf("Failed to prepare table creation statement: %v", err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func getCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	cotacao, err := getCotacao()
	if err != nil {
		// Log the error to terminal
		log.Printf("Error in getCotacao: %v\n", err)

		// Set the response headers and status code
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		// Send a JSON response to the client
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Internal Server Error",
			"message": fmt.Sprintf("Failed to get cotacao: %v", err),
		})
		return 
	}

	// Save the bid value to SQLite with a 10ms timeout
	err = saveCotacaoToDB(cotacao.Bid)
	if err != nil {
		// Log the error to terminal
		log.Printf("Error saving to database: %v\n", err)

		// Set the response headers and status code
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		// Send a JSON response to the client
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Internal Server Error",
			"message": fmt.Sprintf("Failed to save cotacao to database: %v", err),
		})
		return
	}

	// Create a new map to hold just the bid value
	bid := map[string]string{"bid": cotacao.Bid}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid)
}

func getCotacao() (*Cotacao, error) {
	// Create a context with a 200ms timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("API request timed out after 200ms")
		}

		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from API: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseMap map[string]Cotacao
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		return nil, err
	}

	// Check if the key "USDBRL" exists in the map
	cotacaoValue, exists := responseMap["USDBRL"]
	if !exists {
		return nil, fmt.Errorf("key USDBRL does not exist in the response map")
	}

	return &cotacaoValue, nil
}

func saveCotacaoToDB(bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		log.Println("Database operation timed out after 10ms")
	}
	return err
}
