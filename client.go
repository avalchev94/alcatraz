package alcatraz

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/avalchev94/alcatraz/pb"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientConfig struct {
	Host            string
	MonitorFolder   string
	MonitorInterval time.Duration
	Certificates    CertFiles
	ParallelUploads int
	ChunkSize       int
	LogLevel        string
}

type Client struct {
	ClientConfig
	cli   pb.AlcatrazClient
	creds credentials.TransportCredentials
}

func NewClient(config ClientConfig) (*Client, error) {
	// check if folder exists
	if _, err := os.Stat(config.MonitorFolder); os.IsNotExist(err) {
		return nil, fmt.Errorf("folder %q does not exist", config.MonitorFolder)
	}

	// set logging level
	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse log level: %v", err)
	}
	logrus.SetLevel(lvl)

	// Load the client certificate
	certificate, err := config.Certificates.getCertificate()
	if err != nil {
		return nil, fmt.Errorf("could not get certificate: %v", err)
	}

	// Get a certificate pool with the certificate authority
	certPool, err := config.Certificates.getCertAuthPool()
	if err != nil {
		return nil, fmt.Errorf("could not get cert auth pool: %v", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	return &Client{
		ClientConfig: config,
		cli:          nil,
		creds:        creds,
	}, nil
}

func (c *Client) Run(ctx context.Context) error {
	// open connection to the server
	conn, err := grpc.Dial(c.Host, grpc.WithTransportCredentials(c.creds))
	if err != nil {
		return fmt.Errorf("failed to dial server on host %q: %v", c.Host, err)
	}
	defer conn.Close()

	c.cli = pb.NewAlcatrazClient(conn)
	log.Infof("Opened connection with %s", c.Host)

	// create channels for the monitoring and uploading goroutines
	upload := make(chan string, 100)
	uploaded := make(chan string, 100)
	failed := make(chan string, 100)

	// run 1 monitor goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		c.monitor(ctx, upload, uploaded, failed)
		wg.Done()
	}()

	// run the specified number of goroutines for uploading files
	for i := 0; i < c.ParallelUploads; i++ {
		wg.Add(1)
		go func() {
			c.upload(ctx, upload, uploaded, failed)
			wg.Done()
		}()
	}

	// if context is canceled, wait goroutines to exit and return
	<-ctx.Done()
	log.Info("Gracefully stoping client...")
	wg.Wait()
	log.Info("Client stopped")

	return nil
}

func (c *Client) monitor(ctx context.Context, upload chan<- string, uploaded, failed <-chan string) {
	uploadingFiles := map[string]struct{}{}

	timer := time.NewTimer(1)
	for {
		select {
		case <-ctx.Done():
			return
		case file := <-uploaded:
			if err := os.Remove(file); err != nil {
				log.Errorf("Failed to delete file %q. Error: %v", file, err)
			}
			delete(uploadingFiles, file)
		case file := <-failed:
			delete(uploadingFiles, file)
		case <-timer.C:
			err := filepath.Walk(c.MonitorFolder, func(filepath string, fileinfo os.FileInfo, err error) error {
				if fileinfo.IsDir() {
					if filepath != c.MonitorFolder {
						c.removeIfEmpty(filepath)
					}
					return nil
				}

				if _, ok := uploadingFiles[filepath]; !ok {
					// if upload channel is full, we don't want to block monitor goroutine
					go func() { upload <- filepath }()
					uploadingFiles[filepath] = struct{}{}
				}
				return err
			})
			if err != nil {
				log.Errorf("Failed to monitor folder: %v", err)
			}

			timer.Reset(5 * time.Second)
		}
	}
}

func (c *Client) upload(ctx context.Context, upload <-chan string, uploaded, failed chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case file := <-upload:
			log.Debugf("Uploading %q...", file)
			if err := c.uploadFile(ctx, file); err != nil {
				failed <- file
				log.Errorf("Failed to upload %q. Error: %v", file, err)
			} else {
				uploaded <- file
				log.Debugf("File %q was uploaded.", file)
			}
		}
	}
}

func (c *Client) uploadFile(ctx context.Context, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	// create stream for uploading the file
	stream, err := c.cli.UploadFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %v", err)
	}

	// stream the file
	if err := c.streamFile(file, stream); err != nil {
		// even if the streaming is not successful, we want to close and recieve message from the server
		// the server might have some useful error, that he sent for the client
		if _, closeErr := stream.CloseAndRecv(); closeErr != nil {
			return fmt.Errorf("failed stream the file, stream error: %v, close error: %v", err, closeErr)
		}
		return fmt.Errorf("failed to stream the file: %v", err)
	}

	// close the stream
	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("failed to close the stream: %v", err)
	}

	return nil
}

func (c *Client) streamFile(file *os.File, stream pb.Alcatraz_UploadFileClient) error {
	// always send the filename first
	err := stream.Send(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Name{
			Name: strings.TrimPrefix(file.Name(), c.MonitorFolder),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send file's name: %v", err)
	}

	// send the file data chunk by chunk
	hash := sha256.New()
	for {
		chunk := make([]byte, c.ChunkSize)
		if _, err := file.Read(chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read chunk from file: %v", err)
		}

		if _, err := hash.Write(chunk); err != nil {
			return fmt.Errorf("failed to write to hash: %v", err)
		}

		err = stream.Send(&pb.UploadRequest{
			TestOneof: &pb.UploadRequest_Chunk{
				Chunk: chunk,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to send chunck: %v", err)
		}
	}

	// always send the hash last
	err = stream.Send(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Hash{
			Hash: hex.EncodeToString(hash.Sum(nil)),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send file's hash: %v", err)
	}

	return nil
}

func (c *Client) removeIfEmpty(dirname string) error {
	content, err := ioutil.ReadDir(dirname)
	if err != nil {
		return fmt.Errorf("failed to read dir: %v", err)
	}

	if len(content) > 0 {
		return nil
	}

	return os.Remove(dirname)
}
