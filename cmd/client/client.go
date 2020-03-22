package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/avalchev94/alcatraz"
)

func main() {
	var (
		host     = flag.String("host", "localhost:8080", "the address of the Alcatraz server")
		cert     = flag.String("crt", "", "path to the client certificate")
		key      = flag.String("key", "", "path to client private key")
		ca       = flag.String("ca", "", "path to the Certificate Authority certificate")
		parallel = flag.Int("parallel", 30, "maximum number of parallely processed files")
		interval = flag.Duration("interval", 5*time.Second, "folder monitoring interval")
		chunk    = flag.Int("chunk", 32000, "size(in bytes) of the chunk unit(files are divided to chunks when uploaded)")
		log      = flag.String("log", "info", "log level: info or debug")
	)

	if len(os.Args) < 2 {
		fmt.Printf("Usage:\n\talcatraz path_to_folder\n")
		os.Exit(1)
	}

	flag.CommandLine.Parse(os.Args[2:])

	cfg := alcatraz.ClientConfig{
		Host:            *host,
		MonitorFolder:   os.Args[1],
		MonitorInterval: *interval,
		Certificates: alcatraz.CertFiles{
			Certificate: *cert,
			Key:         *key,
			CertAuth:    *ca,
		},
		ParallelUploads: *parallel,
		ChunkSize:       *chunk,
		LogLevel:        *log,
	}

	client, err := alcatraz.NewClient(cfg)
	if err != nil {
		fmt.Printf("Failed to create Alcatraz client: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		if err := client.Run(ctx); err != nil {
			fmt.Printf("Failed to run Alcatraz client: %v\n", err)
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
