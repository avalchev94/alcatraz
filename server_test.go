package alcatraz

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/avalchev94/alcatraz/pb"
	"github.com/avalchev94/alcatraz/pb/mock"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/testdata"
)

func TestUploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server := Server{
		ServerConfig: ServerConfig{
			StoragePath: "storage",
			AllowedClients: map[string]bool{
				"Asenski": true,
			},
		},
	}

	stream := mock.NewMockAlcatraz_UploadFileServer(ctrl)

	// always return good peer when Context called
	clientName := "Asenski"
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					&x509.Certificate{
						Subject: pkix.Name{
							CommonName: clientName,
						},
					},
				},
			},
		},
	})
	stream.EXPECT().Context().AnyTimes().Return(ctx)

	// test with error from Recv
	stream.EXPECT().Recv().Times(1).Return(nil, errors.New("bad call"))
	if err := server.UploadFile(stream); err == nil {
		t.Errorf("expected error, but got nil")
	} else if grpc.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, but got %v", grpc.Code(err))
	}

	// test with good return from Recv
	stream.EXPECT().Recv().Times(1).Return(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Chunk{
			Chunk: []byte("this test"),
		},
	}, nil)

	// but wrong hash meta
	stream.EXPECT().Recv().Times(1).Return(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Metadata{
			Metadata: &pb.Metadata{
				Hash: "1234",
			},
		},
	}, nil)
	if err := server.UploadFile(stream); err == nil {
		t.Errorf("expected error, but got nil")
	} else if grpc.Code(err) != codes.DataLoss {
		t.Errorf("expected DataLoss, but got %v", grpc.Code(err))
	}

	// test with good return from Recv
	stream.EXPECT().Recv().Times(1).Return(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Chunk{
			Chunk: []byte("this test"),
		},
	}, nil)

	stream.EXPECT().Recv().Times(1).Return(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Chunk{
			Chunk: []byte(" is awsome"),
		},
	}, nil)

	// and good hash
	stream.EXPECT().Recv().Times(1).Return(&pb.UploadRequest{
		TestOneof: &pb.UploadRequest_Metadata{
			Metadata: &pb.Metadata{
				Name: "file.txt",
				Hash: "3d6c07b31fef053b227cdd9256e8eb7d314766bb2360ed3d3c562c57ad0ba696",
			},
		},
	}, nil)

	stream.EXPECT().SendAndClose(&empty.Empty{}).Return(nil)

	if err := server.UploadFile(stream); err != nil {
		t.Errorf("expected nil error, but got %v", err)
	}

	// verify file was created
	_, err := os.Stat(filepath.Join(server.StoragePath, clientName, "file.txt"))
	if err != nil {
		t.Errorf("file does not exist")
	}

	// clean the file
	os.RemoveAll(server.StoragePath)
}

func TestNewServer(t *testing.T) {
	// bad logging level
	config := ServerConfig{
		LogLevel: "asdasd",
	}
	if _, err := NewServer(config); err == nil {
		t.Error("logging level is bad, error should be returned")
	}

	// bad certificates
	config = ServerConfig{
		LogLevel: "debug",
		Certificates: CertFiles{
			Certificate: "cert.crt",
			Key:         "cert.key",
		},
	}
	if _, err := NewServer(config); err == nil {
		t.Error("cert paths are bad, error should be returned")
	}

	// bad certificate auth
	config = ServerConfig{
		LogLevel: "debug",
		Certificates: CertFiles{
			Certificate: testdata.Path("server1.pem"),
			Key:         testdata.Path("server1.key"),
			CertAuth:    "asdsad.crt",
		},
	}
	if _, err := NewServer(config); err == nil {
		t.Error("cert auth is bad, error should be returned")
	}

	// success
	config = ServerConfig{
		LogLevel: "debug",
		Certificates: CertFiles{
			Certificate: testdata.Path("server1.pem"),
			Key:         testdata.Path("server1.key"),
			CertAuth:    testdata.Path("ca.pem"),
		},
		StoragePath: "storage",
	}
	if _, err := NewServer(config); err != nil {
		t.Errorf("expected success, got %v", err)
	}

	if _, err := os.Stat(config.StoragePath); err != nil {
		t.Error("storage path should have been created")
	}

	os.RemoveAll(config.StoragePath)
}
