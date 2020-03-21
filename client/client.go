package main

import (
	"context"
	"log"
	"time"

	"github.com/avalchev94/alcatraz/pb"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAlcatrazClient(conn)

	stream, err := client.UploadFile(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	err = stream.Send(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Metadata{
			Metadata: &pb.Metadata{
				Name: "file1.txt",
				Size: 10,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	stream.Send(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Chunk{
			Chunk: []byte("Geri e"),
		},
	})

	stream.Send(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Chunk{
			Chunk: []byte(" gei"),
		},
	})

	stream.CloseSend()

	time.Sleep(10 * time.Second)
}
