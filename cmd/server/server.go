package main

import (
	"context"
	"fmt"
	"os"

	"github.com/avalchev94/alcatraz"
)

func main() {
	cfg := alcatraz.ServerConfig{
		Port:        8080,
		StoragePath: "storage/",
		Certificates: alcatraz.CertFiles{
			Certificate: "../../certs/localhost.crt",
			Key:         "../../certs/localhost.key",
			CertAuth:    "../../certs/CertAuth.crt",
		},
		AllowedClients: map[string]bool{
			"Alice": true,
		},
	}

	server := alcatraz.NewServer(cfg)
	if err := server.Run(context.Background()); err != nil {
		fmt.Printf("Failed to run Alcatraz server: %v\n", err)
		os.Exit(1)
	}
}
