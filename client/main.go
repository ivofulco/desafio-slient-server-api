package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return
	} else if resp.StatusCode >= 500 {
		// Handle HTTP errors
		fmt.Printf("Server error: %v. An error occurred while processing the request\n", resp.StatusCode)
	} else {
		// Handle successful response
		// Remember to close the response body
		defer resp.Body.Close()
		// Process the response...

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}

		err = saveToFile(response.Bid)
		if err != nil {
			fmt.Println("Error saving to file:", err)
			return
		}

		fmt.Println("Successfully saved bid value:", response.Bid)
	}

}

// saveToFile saves the bid value to "cotacao.txt"
func saveToFile(content string) error {
	// Open the file in append mode, or create it if it doesn't exist
	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Append content to the file in the format "Dolar: {value}"
	_, err = io.WriteString(file, fmt.Sprintf("Dolar: %s\n", content))
	if err != nil {
		return err
	}
	return nil
}
