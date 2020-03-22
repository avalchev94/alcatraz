package main

import (
	"context"
	"fmt"
	"os"

	"github.com/avalchev94/alcatraz"
	"github.com/sirupsen/logrus"
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
			"Malcolm": true,
			"Dewey":   true,
			"Reese":   true,
		},
	}

	logrus.SetLevel(logrus.DebugLevel)

	server := alcatraz.NewServer(cfg)
	if err := server.Run(context.Background()); err != nil {
		fmt.Printf("Failed to run Alcatraz server: %v\n", err)
		os.Exit(1)
	}
}
