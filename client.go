package alcatraz

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/avalchev94/alcatraz/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Client struct {
	host   string
	folder string
	cli    pb.AlcatrazClient
}

func NewClient(host string, folder string) *Client {
	return &Client{
		host:   host,
		folder: folder,
		cli:    nil,
	}
}

func (c *Client) Run(ctx context.Context) error {
	// check if folder exists
	if _, err := os.Stat(c.folder); os.IsNotExist(err) {
		return fmt.Errorf("folder %q does not exist", c.folder)
	}

	// we will work only with files inside the folder, so change working dir
	if err := os.Chdir(c.folder); err != nil {
		return fmt.Errorf("failed to change working directory: %v", err)
	}

	// open connection to the server
	conn, err := grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to dial server on host %q: %v", c.host, err)
	}
	defer conn.Close()

	c.cli = pb.NewAlcatrazClient(conn)
	log.Infof("Opened connection with %s", c.host)

	upload := make(chan string, 50)
	uploaded := make(chan string, 50)
	failed := make(chan string, 50)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		c.monitor(ctx, upload, uploaded, failed)
		wg.Add(1)
	}()
	go func() {
		c.upload(ctx, upload, uploaded, failed)
		wg.Add(1)
	}()

	// if context is canceled, wait goroutines to exit and return
	<-ctx.Done()
	wg.Wait()

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
			if err := c.removeFile(file); err != nil {
				log.Errorf("Failed to delete file %q. Error: %v", file, err)
			}
			delete(uploadingFiles, file)
		case file := <-failed:
			delete(uploadingFiles, file)
		case <-timer.C:
			err := filepath.Walk(".", func(filepath string, fileinfo os.FileInfo, err error) error {
				if fileinfo.IsDir() {
					return nil
				}

				if _, ok := uploadingFiles[filepath]; !ok {
					upload <- filepath
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
			go func() {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				log.Debugf("Uploading %q...", file)
				if err := c.uploadFile(ctx, file); err != nil {
					failed <- file
					log.Errorf("Failed to upload %q. Error: %v", file, err)
				} else {
					uploaded <- file
					log.Debugf("File %q was uploaded.", file)
				}
			}()
		}
	}
}

func (c *Client) uploadFile(ctx context.Context, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file's stat: %v", err)
	}

	// create stream for uploading the file
	stream, err := c.cli.UploadFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %v", err)
	}

	// always send the metadata first
	err = stream.Send(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Metadata{
			Metadata: &pb.Metadata{
				Name: filename,
				Size: fileinfo.Size(),
				// hash
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send file's metadata: %v", err)
	}

	// send the file data chunk by chunk
	for {
		chunk := make([]byte, 64) // move me out of here
		if _, err := file.Read(chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read chunk from file: %v", err)
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

	// close the stream
	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("failed to close the stream: %v", err)
	}

	return nil
}

func (c *Client) removeFile(filename string) error {
	if err := os.Remove(filename); err != nil {
		return err
	}

	if dir := filepath.Dir(filename); dir != "" {
		if err := os.Remove(dir); err == nil {
			err = nil
		}
	}

	return nil
}
