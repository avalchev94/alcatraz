package alcatraz

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/avalchev94/alcatraz/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

type ServerConfig struct {
	Port           int
	StoragePath    string
	Certificates   CertFiles
	AllowedClients map[string]bool
	LogLevel       string
}

type Server struct {
	ServerConfig
	creds credentials.TransportCredentials
}

func NewServer(config ServerConfig) (*Server, error) {
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
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	// finally, create the storage folder, if not exist
	if _, err := os.Stat(config.StoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(config.StoragePath, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create storage folder %q: %v", config.StoragePath, err)
		}
	}

	return &Server{
		ServerConfig: config,
		creds:        creds,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen to port %d: %v", s.Port, err)
	}

	server := grpc.NewServer(grpc.Creds(s.creds), grpc.StreamInterceptor(s.authClient))
	pb.RegisterAlcatrazServer(server, s)

	log.Infof("Server listening on port %d...", s.Port)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()

	<-ctx.Done()

	log.Info("Gracefully stoping server...")
	server.GracefulStop()
	log.Info("Server stopped")

	return nil
}

func (s *Server) authClient(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	name, err := getCommonNameFromCtx(ss.Context())
	if err != nil {
		return grpc.Errorf(codes.Unauthenticated, "failed to retrieve client common name: %v", err)
	}

	if !s.AllowedClients[name] {
		return grpc.Errorf(codes.Unauthenticated, "common name is not allowed")
	}

	return handler(srv, ss)
}

func (s *Server) UploadFile(stream pb.Alcatraz_UploadFileServer) error {
	client, _ := getCommonNameFromCtx(stream.Context())

	filename, err := s.recvFilename(stream)
	if err != nil {
		return grpc.Errorf(codes.InvalidArgument, "failed to recieve metadata: %v", err)
	}
	log.Debugf("Client [%s]: started file %q upload..", client, filename)

	// create all sub-directories for the file
	fullname := filepath.Join(s.StoragePath, client, filename)
	dir := filepath.Dir(fullname)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return grpc.Errorf(codes.Internal, "failed to create directories: %v", err)
		}
	}

	file, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return grpc.Errorf(codes.Internal, "failed to create temp file: %v", err)
	}

	// that's a little bit tricky, if success stays false, the file will be deleted
	// if it's changed to true, file will be closed
	success := false
	defer func() {
		if success {
			file.Close()
		} else {
			os.Remove(file.Name())
		}
	}()

	hash := sha256.New()
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Errorf("Client [%s]: file upload failed on Recv with error: %v", client, err)
			return grpc.Errorf(codes.InvalidArgument, "failed to recieve msg from stream: %v", err)
		}

		chunk := msg.GetChunk()
		if chunk == nil {
			// hash is always the last message, verify it
			if hex.EncodeToString(hash.Sum(nil)) != msg.GetHash() {
				log.Errorf("Client [%s]: file upload failed becase hashes are not equal", client)
				return grpc.Errorf(codes.DataLoss, "hashes are not equal")
			}
			break
		}

		if _, err := hash.Write(chunk); err != nil {
			log.Errorf("Client [%s]: file upload failed on hash write with error: %v", client, err)
			return grpc.Errorf(codes.Internal, "failed to write to hash: %v", err)
		}

		if _, err := file.Write(chunk); err != nil {
			log.Errorf("Client [%s]: file upload failed on file write with error: %v", client, err)
			return grpc.Errorf(codes.Internal, "failed to write to hash: %v", err)
		}
	}

	// change success to true, we want to close the file, not delete it!
	success = true
	log.Debugf("Client [%s]: file with name %q was uploaded", client, filename)

	return stream.SendAndClose(&empty.Empty{})
}

func (s *Server) recvFilename(stream pb.Alcatraz_UploadFileServer) (string, error) {
	msg, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("failed to recieve msg from stream: %v", err)
	}

	if msg.GetName() == "" {
		return "", fmt.Errorf("expected filename")
	}

	return msg.GetName(), nil
}
