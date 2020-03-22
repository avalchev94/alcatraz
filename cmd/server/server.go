package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/avalchev94/alcatraz"
)

func main() {
	var (
		port    = flag.Int("port", 8080, "port on which Alcatraz server will listen")
		storage = flag.String("storage", "storage", "path to the storage folder")
		cert    = flag.String("crt", "", "path to the client certificate")
		key     = flag.String("key", "", "path to client private key")
		ca      = flag.String("ca", "", "path to the Certificate Authority certificate")
		log     = flag.String("log", "info", "log level: info or debug")
	)
	flag.Parse()

	cfg := alcatraz.ServerConfig{
		Port:        *port,
		StoragePath: *storage,
		Certificates: alcatraz.CertFiles{
			Certificate: *cert,
			Key:         *key,
			CertAuth:    *ca,
		},
		AllowedClients: map[string]bool{
			"Malcolm": true,
			"Dewey":   true,
			"Reese":   true,
		},
		LogLevel: *log,
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		if err := alcatraz.NewServer(cfg).Run(ctx); err != nil {
			fmt.Printf("Failed to run Alcatraz server: %v\n", err)
			os.Exit(1)
		}
		wg.Done()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	cancel()
	wg.Wait()
}
