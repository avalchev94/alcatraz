package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/avalchev94/alcatraz/pb"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAlcatrazServer(s, &Server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type Server struct {
}

func (s *Server) UploadFile(stream pb.Alcatraz_UploadFileServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	meta := msg.GetMetadata()
	if meta == nil {
		return fmt.Errorf("First message should be metadata of the file")
	}

	buffer := make([]byte, 0, meta.GetSize())
	for done := false; !done; {
		select {
		case <-stream.Context().Done():
			log.Fatal("cancel")
			// cancel?
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				done = true
				break
			}

			if err != nil {
				return err
			}

			buffer = append(buffer, msg.GetChunk()...)
		}
	}

	fmt.Println(string(buffer))

	return stream.SendAndClose(&pb.UploadResponse{
		Success: true,
		Message: "",
	})
}
