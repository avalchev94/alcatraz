package alcatraz

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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
}

func NewServer(config ServerConfig) *Server {
	return &Server{
		ServerConfig: config,
	}
}

func (s *Server) Run(ctx context.Context) error {
	if _, err := os.Stat(s.StoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.StoragePath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create storage folder %q: %v", s.StoragePath, err)
		}
	}

	// set logging level
	lvl, err := logrus.ParseLevel(s.LogLevel)
	if err != nil {
		return fmt.Errorf("couldn't parse log level: %v", err)
	}
	logrus.SetLevel(lvl)

	// Load the client certificate
	certificate, err := s.Certificates.getCertificate()
	if err != nil {
		return fmt.Errorf("could not get certificate: %v", err)
	}

	// Get a certificate pool with the certificate authority
	certPool, err := s.Certificates.getCertAuthPool()
	if err != nil {
		return fmt.Errorf("could not get cert auth pool: %v", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen to port %d: %v", s.Port, err)
	}

	server := grpc.NewServer(grpc.Creds(creds), grpc.StreamInterceptor(s.authClient))
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
	log.Debugf("Client [%s]: started file upload..", client)

	var (
		meta   *pb.Metadata
		buffer = []byte{}
		hash   = sha256.New()
	)

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Errorf("Client [%s]: file upload failed on Recv with error: %v", client, err)
			return grpc.Errorf(codes.InvalidArgument, "failed to recieve msg from stream: %v", err)
		}

		chunk := msg.GetChunk()
		if chunk == nil {
			meta = msg.GetMetadata()
			break
		}

		if _, err := hash.Write(chunk); err != nil {
			log.Errorf("Client [%s]: file upload failed on hash write with error: %v", client, err)
			return grpc.Errorf(codes.Internal, "failed to write to hash: %v", err)
		}

		buffer = append(buffer, chunk...)
	}

	if hex.EncodeToString(hash.Sum(nil)) != meta.GetHash() {
		log.Errorf("Client [%s]: file upload failed becase hashes are not equal", client)
		return grpc.Errorf(codes.InvalidArgument, "hashes are not equal")
	}

	if err := s.writeFile(client, meta.GetName(), buffer); err != nil {
		log.Errorf("Client [%s]: file upload failed on write file with error: %v", client, err)
		return grpc.Errorf(codes.Internal, "failed to create the file: %v", err)
	}
	log.Debugf("Client [%s]: file with name %q and size %d was uploaded", client, meta.GetName(), meta.GetSize())

	return stream.SendAndClose(&empty.Empty{})
}

func (s *Server) writeFile(client, filename string, buffer []byte) error {
	fullpath := filepath.Join(s.StoragePath, client, filename)

	if err := os.MkdirAll(filepath.Dir(fullpath), os.ModePerm); err != nil {
		return fmt.Errorf("failed create, prior to file, directories: %v", err)
	}

	if err := ioutil.WriteFile(fullpath, buffer, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
