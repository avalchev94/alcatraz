package main

import (
	"context"
	"fmt"
	"os"

	"github.com/avalchev94/alcatraz"
)

func main() {
	client := alcatraz.NewClient("localhost:8080", "upload/")
	if err := client.Run(context.Background()); err != nil {
		fmt.Printf("Failed to run Alcatraz client: %v\n", err)
		os.Exit(1)
	}
}
