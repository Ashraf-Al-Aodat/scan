package analysis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"scan/internal/clients"
	"scan/internal/prompts"
	"scan/internal/utils"
)

// PerformAnalysis analyzes the provided files using the LLM client.
func PerformAnalysis(client *clients.Client, files []string) {
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan string, len(files))

	log.Println("AI analysis started...")

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			start := time.Now()
			// Step 1: Read file
			fileData, err := utils.ReadFile(file)
			if err != nil {
				results <- fmt.Sprintf("Failed to read file %s: %v", file, err)
				return
			}

			// Step 2: Load role prompt
			systemRole, err := prompts.GetSystemRole("security")
			if err != nil {
				results <- fmt.Sprintf("Failed to load role prompt for %s: %v\n", file, err)
				return
			}

			// Step 3: Format and send prompt to LLM
			prompt := fmt.Sprintf(systemRole, fileData)
			response, err := client.GenerateResponse(ctx, prompt)
			if err != nil {
				results <- fmt.Sprintf("Failed to generate LLM response for %s: %v\n", file, err)
				return
			}

			duration := time.Since(start)
			results <- fmt.Sprintf("Analysis for %s (Time taken: %v): %s\n", file, duration, response)
		}(file)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Print(result)
	}

	log.Println("AI analysis completed!")
}
