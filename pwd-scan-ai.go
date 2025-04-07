package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Start overall timer
	overallStart := time.Now()
	var envs envFlag

	path := flag.String("p", "", "Specify the path of the local repo to scan.")
	url := flag.String("h", "", "Specify the host url.")
	flag.Var(&envs, "e", "Specify environment variable names. Can be used multiple times.")

	// Parse flags
	flag.Parse()

	// Validate required flags
	if *path == "" || *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	host, llmUrl, err := formatHost(*url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Retrieve environment variable values
	headers := make(map[string]string)
	for _, envVar := range envs {
		value := os.Getenv(envVar)
		if value == "" {
			log.Fatalf("Environment variable %s is not set.", envVar)
		}
		headers[strings.ReplaceAll(envVar, "_", "-")] = value
	}

	for key, value := range headers {
		log.Printf("%s, %s", key, value)
	}

	files, err := getFiles(path)
	if err != nil {
		log.Fatal(err)
	}

	// Configure the OpenAI client to use the ELI endpoint
	// Set up the OpenAI-compatible LLM with custom endpoint and HTTP client
	client := newCustomClient(host, headers)
	llm, err := openai.New(
		openai.WithBaseURL(llmUrl),
		openai.WithModel("Mistral-24b"),
		openai.WithHTTPClient(client),
	)
	if err != nil {
		panic(err)
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
			prompt := fmt.Sprintf("You are a security expert. Analyze the following file for sensitive data lie pwd, apikeys, etc.\nif you find anything respsned with either: found: {description of the issue} or none.\n\n%s", fileData)

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
		if !info.IsDir() && strings.Contains(path, "go") {
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

// newCustomClient creates an HTTP client with custom headers.
func newCustomClient(host string, headers map[string]string) *http.Client {
	return &http.Client{
		Transport: &customTransport{
			headers: headers,
			host:    host,
			rt:      http.DefaultTransport,
		},
	}
}

type customTransport struct {
	host    string
	headers map[string]string
	rt      http.RoundTripper
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "LangChainGo/1.0")

	// Fix: remove default OpenAI "application/x-www-form-urlencoded"
	req.Header.Del("Content-Type")

	// Fix: override Host header if required (optional)
	req.Host = c.host

	return c.rt.RoundTrip(req)
}

// envFlag is a custom flag type to collect environment variables.
type envFlag []string

// String returns the string representation of the environment variable names.
func (e *envFlag) String() string {
	return strings.Join(*e, ", ")
}

// Set appends a new environment variable name to the slice.
func (e *envFlag) Set(value string) error {
	*e = append(*e, value)
	return nil
}

func formatHost(userURL string) (string, string, error) {
	// Ensure the URL has a scheme; default to "https" if missing
	if !strings.HasPrefix(userURL, "http://") && !strings.HasPrefix(userURL, "https://") {
		userURL = "https://" + userURL
	}

	parsedURL, err := url.Parse(userURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	host := parsedURL.Hostname()
	llmURL := fmt.Sprintf("https://%s/api/openai/v1", host)

	return host, llmURL, nil
}
