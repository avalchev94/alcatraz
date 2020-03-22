package main

import (
	"context"
	"fmt"
	"os"

	"github.com/avalchev94/alcatraz"
)

func main() {
	cfg := alcatraz.ClientConfig{
		Host:          "localhost:8080",
		MonitorFolder: "upload/",
		Certificates: alcatraz.CertFiles{
			Certificate: "../../certs/Reese.crt",
			Key:         "../../certs/Reese.key",
			CertAuth:    "../../certs/CertAuth.crt",
		},
	}

	client := alcatraz.NewClient(cfg)
	if err := client.Run(context.Background()); err != nil {
		fmt.Printf("Failed to run Alcatraz client: %v\n", err)
		os.Exit(1)
	}
}
