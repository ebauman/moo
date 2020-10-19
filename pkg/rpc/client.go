package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

func LoadTLSCredentials(caCert string) (credentials.TransportCredentials, error) {
	var certPool *x509.CertPool
	if len(caCert) > 0 {
		certPool = x509.NewCertPool()

		serverCA, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, err
		}

		if !certPool.AppendCertsFromPEM(serverCA) {
			return nil, fmt.Errorf("failed to add ca certificate to cert pool")
		}
	} else {
		certPool, _ = x509.SystemCertPool()
		if certPool == nil {
			certPool = x509.NewCertPool()
		}
	}

	config := &tls.Config {
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func setupGrpcClient(hostname string, insecure bool, cacert string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	if insecure {
		conn, err = grpc.Dial(hostname, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
	} else {
		tlsConfig, err := LoadTLSCredentials(cacert)
		if err != nil {
			return nil, err
		}
		conn, err = grpc.Dial(hostname, grpc.WithTransportCredentials(tlsConfig))
	}

	return conn, nil
}

func SetupMooClient(hostname string, insecure bool, cacert string) (MooClient, error) {
	client, err := setupGrpcClient(hostname, insecure, cacert)
	if err != nil {
		return nil, err
	}

	return NewMooClient(client), nil
}

func SetupRulesClient(hostname string, insecure bool, cacert string) (RulesClient, error) {
	client, err := setupGrpcClient(hostname, insecure, cacert)
	if err != nil {
		return nil, err
	}

	return NewRulesClient(client), nil
}

func SetupClients(hostname string, insecure bool, cacert string) (MooClient, RulesClient, error) {
	client, err := setupGrpcClient(hostname, insecure, cacert)
	if err != nil {
		return nil, nil, err
	}

	mooClient := NewMooClient(client)
	rulesClient := NewRulesClient(client)

	return mooClient, rulesClient, nil
}
