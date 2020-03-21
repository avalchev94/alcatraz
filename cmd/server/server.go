package main

import (
	"context"
	"fmt"
	"os"

	"github.com/avalchev94/alcatraz"
)

func main() {
	server := alcatraz.NewServer(8080, "storage")
	if err := server.Run(context.Background()); err != nil {
		fmt.Printf("Failed to run Alcatraz server: %v\n", err)
		os.Exit(1)
	}
}
