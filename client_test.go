package alcatraz

import (
	"os"
	"testing"

	"google.golang.org/grpc/testdata"
)

func TestNewClient(t *testing.T) {
	config := ClientConfig{
		MonitorFolder: "upload",
	}
	if _, err := NewClient(config); err == nil {
		t.Error("monitor folder does not exist, error should be returned")
	}

	monitorFolder := "upload"
	os.MkdirAll(monitorFolder, os.ModePerm)
	defer os.RemoveAll(monitorFolder)

	// bad log level
	config = ClientConfig{
		MonitorFolder: monitorFolder,
		LogLevel:      "asdasd",
	}
	if _, err := NewClient(config); err == nil {
		t.Error("bad log level, error should be returned")
	}

	// bad certificates
	config = ClientConfig{
		MonitorFolder: monitorFolder,
		LogLevel:      "debug",
		Certificates: CertFiles{
			Certificate: "lals",
			Key:         "asdas",
		},
	}
	if _, err := NewClient(config); err == nil {
		t.Error("bad certificate, error should be returned")
	}

	// bad certificate auth
	config = ClientConfig{
		MonitorFolder: monitorFolder,
		LogLevel:      "debug",
		Certificates: CertFiles{
			Certificate: testdata.Path("server1.pem"),
			Key:         testdata.Path("server1.key"),
			CertAuth:    "asdsad.crt",
		},
	}
	if _, err := NewClient(config); err == nil {
		t.Error("bad certificate auth, error should be returned")
	}

	// success
	config = ClientConfig{
		MonitorFolder: monitorFolder,
		LogLevel:      "debug",
		Certificates: CertFiles{
			Certificate: testdata.Path("server1.pem"),
			Key:         testdata.Path("server1.key"),
			CertAuth:    testdata.Path("ca.pem"),
		},
	}
	if _, err := NewClient(config); err != nil {
		t.Errorf("expected success, got %v", err)
	}
}
