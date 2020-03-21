package alcatraz

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/avalchev94/alcatraz/pb"
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

	// open connection to the server
	conn, err := grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to dial server on host %q: %v", c.host, err)
	}
	defer conn.Close()
	c.cli = pb.NewAlcatrazClient(conn)

	upload := make(chan string, 100)
	uploaded := make(chan string, 100)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		c.monitor(ctx, upload, uploaded)
		wg.Add(1)
	}()
	go func() {
		c.upload(ctx, upload, uploaded)
		wg.Add(1)
	}()

	// if context is canceled, wait goroutines to exit and return
	<-ctx.Done()
	wg.Wait()

	return nil
}

func (c *Client) monitor(ctx context.Context, upload chan<- string, uploaded <-chan string) {
	uploadingFiles := map[string]struct{}{}

	timer := time.NewTimer(1)
	for {
		select {
		case <-ctx.Done():
			return
		case file := <-uploaded:
			if err := os.Remove(file); err != nil {
				log.Printf("failed to delete file: %v", err)
			}
			delete(uploadingFiles, file)
		case <-timer.C:
			err := filepath.Walk(c.folder, func(path string, _ os.FileInfo, err error) error {
				if path == c.folder {
					return nil
				}

				if _, ok := uploadingFiles[path]; !ok {
					upload <- path
					uploadingFiles[path] = struct{}{}
				}
				return err
			})
			if err != nil {
				log.Fatal(err)
			}

			timer.Reset(5 * time.Second)
		}
	}
}

func (c *Client) upload(ctx context.Context, upload <-chan string, uploaded chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case file := <-upload:
			go func() {
				if err := c.uploadFile(ctx, file); err != nil {
					log.Printf("Failed to upload file %q. Error: %v", file, err)
				} else {
					uploaded <- file
				}
			}()
		}
	}
}

func (c *Client) uploadFile(ctx context.Context, filepath string) error {
	file, err := os.Open(filepath)
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
				Name: fileinfo.Name(),
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
			return fmt.Errorf("failed to send read chunck: %v", err)
		}
	}

	// close the stream
	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("failed to close the stream: %v", err)
	}

	return nil
}
