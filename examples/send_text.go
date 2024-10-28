package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gtfol/textfully-go"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from the same directory as the executable
	if err := godotenv.Load(filepath.Join(".", ".env")); err != nil {
		log.Println("Warning: .env file not found")
	}

	apiKey := os.Getenv("TEXTFULLY_API_KEY")
	if apiKey == "" {
		log.Fatal("TEXTFULLY_API_KEY environment variable not set")
	}

	client := textfully.New(apiKey)

	resp, err := client.Send(
		"+16175555555", // verified phone number
		"Hello, world!",
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Message sent! ID: %s", resp.ID)
}
