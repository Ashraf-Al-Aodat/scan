package main

import (
	"flag"
	"log"
	"os"

	"scan/internal/analysis"
	"scan/internal/clients"
	"scan/internal/config"
	"scan/internal/utils"
)

func main() {
	// Parse command-line flags
	var envNames config.EnvFlag
	path := flag.String("p", "", "Specify the path of the local repo to scan.")
	host := flag.String("h", "", "Specify the host URL.")
	model := flag.String("m", "", "Specify the model name.")
	flag.Var(&envNames, "e", "Specify environment variable names. Can be used multiple times.")

	flag.Parse()

	// Validate required flags
	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *host == "" {
		log.Print("Using default openai host")
	}
	if *host == "" {
		log.Print("Using default model.")
	}

	// Load environment variables
	headers := config.LoadExtraHeadersFromEnvVars(envNames)

	// Initialize LLM client
	client := clients.NewClient(*host, *model, headers)

	// Retrieve files to analyze
	files, err := utils.GetFiles(*path)
	if err != nil {
		log.Fatal(err)
	}

	// Perform analysis
	analysis.PerformAnalysis(client, files)
}
