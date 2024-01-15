package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/kolach/go-factory"

	"transactions/internal/db"
	"transactions/pkg/test"
)

func sendPostRequest(url string, payload interface{}, ch chan<- error) {
	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		ch <- err
		return
	}

	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode == http.StatusOK {
		ch <- nil
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ch <- fmt.Errorf("received non-OK status with: %s", string(bodyBytes))
	}
}

func main() {
	var concurrency, totalRequests int
	var url, delayStr string

	// Parsing command-line arguments
	flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent requests")
	flag.IntVar(&totalRequests, "total", 50, "Total number of requests to send")
	flag.StringVar(&url, "url", "http://example.com", "URL to send requests to")
	flag.StringVar(&delayStr, "delay", "1s", "Delay between requests (e.g., '500ms', '1s')")
	flag.Parse()

	// Parse delay duration
	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		fmt.Println("Invalid delay format:", err)
		return
	}

	var wg sync.WaitGroup
	ch := make(chan error, totalRequests)

	// Start timing
	startTime := time.Now()

	// Create and send requests
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var tr db.Transaction
			test.TransactionFactory.MustSetFields(
				&tr,
				factory.Use(randomdata.FirstName, randomdata.RandomGender).For("UserID"),
			)

			sendPostRequest(url, tr, ch)
		}()

		time.Sleep(delay)

		if i%concurrency == 0 { // ensure concurrency limit
			time.Sleep(time.Millisecond * 10) // slight delay to avoid burst
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(ch)

	// Calculate results
	successCount := 0
	for err := range ch {
		if err == nil {
			successCount++
		} else {
			fmt.Println(err)
		}
	}

	totalTime := time.Since(startTime)
	fmt.Printf("Total successful requests: %d\n", successCount)
	fmt.Printf("Total time taken: %s\n", totalTime)
	if totalTime > 0 {
		fmt.Printf("Requests per second: %.2f\n", float64(successCount)/totalTime.Seconds())
	}
}
