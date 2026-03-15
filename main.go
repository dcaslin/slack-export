package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	resetConf := os.Getenv("RESET_CONF") == "true"
	state, err := NewState(resetConf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing state: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using config path %s\n", state.path)

	token := os.Getenv("SLACK_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_API_TOKEN not set. Please set in environment variables or .env file")
		os.Exit(1)
	}

	exportRoot := os.Getenv("EXPORT_ROOT")
	if exportRoot == "" {
		fmt.Fprintln(os.Stderr, "EXPORT_ROOT not set. Please set in environment variables or .env file")
		os.Exit(1)
	}

	service, err := NewAPIService(
		state,
		token,
		exportRoot,
		os.Getenv("MIN_DATE_ISO"),
		os.Getenv("MAX_DATE_ISO"),
		os.Getenv("EXPORT_JSON") == "true",
		os.Getenv("EXPORT_FILES") == "true",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
		os.Exit(1)
	}

	if err := service.Export(); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing: %v\n", err)
		os.Exit(1)
	}
}
