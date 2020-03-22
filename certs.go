package alcatraz

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type CertFiles struct {
	Certificate string
	Key         string
	CertAuth    string
}

func (c CertFiles) getCertificate() (tls.Certificate, error) {
	certificate, err := tls.LoadX509KeyPair(c.Certificate, c.Key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("could not load client key pair: %s", err)
	}
	return certificate, nil
}

func (c CertFiles) getCertAuthPool() (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(c.CertAuth)
	if err != nil {
		return nil, fmt.Errorf("could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("failed to append ca certs")
	}

	return certPool, nil
}

func getCommonNameFromCtx(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("failed to get peer for context")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", errors.New("failed to cast AuthInfo to TLSInfo")
	}

	if len(tlsAuth.State.PeerCertificates) == 0 {
		return "", errors.New("no certificate found for the peer")
	}

	return tlsAuth.State.PeerCertificates[0].Subject.CommonName, nil
}
