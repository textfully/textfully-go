package main

import (
	"log"
	"os"

	"github.com/gtfol/textfully-go"
)

func main() {
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
