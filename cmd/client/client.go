package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/avalchev94/alcatraz"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := alcatraz.ClientConfig{
		Host:            "localhost:8080",
		MonitorFolder:   "upload/",
		MonitorInterval: 5 * time.Second,
		Certificates: alcatraz.CertFiles{
			Certificate: "../../certs/Reese.crt",
			Key:         "../../certs/Reese.key",
			CertAuth:    "../../certs/CertAuth.crt",
		},
		ParallelUploads: 30,
		ChunkSize:       128,
	}

	logrus.SetLevel(logrus.DebugLevel)

	client := alcatraz.NewClient(cfg)
	if err := client.Run(context.Background()); err != nil {
		fmt.Printf("Failed to run Alcatraz client: %v\n", err)
		os.Exit(1)
	}
}
