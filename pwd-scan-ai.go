package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Start overall timer
	overallStart := time.Now()

	// Parse command-line flag
	path := flag.String("p", "", "Specify the path")
	flag.Parse()
	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Get list of files
	files, err := getFiles(path)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize LLM
	llm, err := openai.New()
	// llm, err = ollama.New(
	// 	ollama.WithModel("deepseek-r1:8b"),
	// 	ollama.WithServerURL("http://localhost:11434"),
	// )
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Sync group and results channel
	var wg sync.WaitGroup
	results := make(chan string, len(files))

	log.Println("AI analysis started...")

	// Process files concurrently
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			// Start timing for this file
			start := time.Now()

			fileData := readFile(file)
			prompt := fmt.Sprintf("You are a security expert. Analyze the following file for sensitive data lie pwd, apikeys, etc.\nif you find anything respsned with either [found , none]\n\n%s", fileData)

			// Generate response
			response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
			if err != nil {
				log.Printf("Error processing file %s: %v", file, err)
				return
			}

			// Measure time taken for this file
			duration := time.Since(start)

			// Send result to channel
			results <- fmt.Sprintf("Analysis for %s (Time taken: %v):\n%s\n", file, duration, response)
		}(file)
	}

	// Close results channel when all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Print results as they arrive
	for result := range results {
		fmt.Print(result)
	}

	// Measure overall execution time
	overallDuration := time.Since(overallStart)
	log.Printf("AI analysis completed! Total time taken: %v", overallDuration)
}

// Retrieves all files from the given path recursively
func getFiles(rootPath *string) ([]string, error) {
	if _, err := os.Stat(*rootPath); err != nil {
		return nil, fmt.Errorf("failed to access root path: %v", err)
	}

	var files []string

	err := filepath.Walk(*rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %v", err)
	}

	return files, nil
}

// Reads file content safely
func readFile(path string) string {
	dat, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(dat)
}
