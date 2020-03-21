package alcatraz

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/avalchev94/alcatraz/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Server struct {
	port    int
	storage string
}

func NewServer(port int, storage string) *Server {
	return &Server{
		port:    port,
		storage: storage,
	}
}

func (s *Server) Run(ctx context.Context) error {
	if _, err := os.Stat(s.storage); os.IsNotExist(err) {
		os.MkdirAll(s.storage, os.ModePerm)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen to port %d: %v", s.port, err)
	}

	server := grpc.NewServer()
	pb.RegisterAlcatrazServer(server, s)
	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *Server) UploadFile(stream pb.Alcatraz_UploadFileServer) error {
	meta, err := s.recvMetadata(stream)
	if err != nil {
		return grpc.Errorf(codes.InvalidArgument, "failed to recv metadata: %v", err)
	}
	log.Printf("file upload: %v", meta)

	buffer := make([]byte, 0, meta.GetSize())
	for {
		chunk, err := s.recvChunk(stream)
		if err != nil {
			return grpc.Errorf(codes.InvalidArgument, "failed to recv chunck: %v", err)
		} else if chunk == nil {
			break
		}

		buffer = append(buffer, chunk...)
	}

	if err := s.writeFile(meta.GetName(), buffer); err != nil {
		return grpc.Errorf(codes.Internal, "failed to create the file: %v", err)
	}

	return stream.SendAndClose(&empty.Empty{})
}

func (s *Server) recvMetadata(stream pb.Alcatraz_UploadFileServer) (*pb.Metadata, error) {
	msg, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to recieve message from stream: %v", err)
	}

	meta := msg.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("message type is not file metada")
	}

	return meta, nil
}

func (s *Server) recvChunk(stream pb.Alcatraz_UploadFileServer) ([]byte, error) {
	msg, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to recieve message from stream: %v", err)
	}

	chunk := msg.GetChunk()
	if chunk == nil {
		return nil, fmt.Errorf("message type is not file chunk")
	}

	return chunk, nil
}

func (s *Server) writeFile(filename string, buffer []byte) error {
	path := filepath.Join(s.storage, filename)
	return ioutil.WriteFile(path, buffer, os.ModePerm)
}
