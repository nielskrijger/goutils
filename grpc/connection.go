package grpc

import (
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServiceConfig struct {
	Address string     `yaml:"address"`
	TLS     *TLSConfig `yaml:"tls"`
}

// TLSConfig contains the TLS configuration.
//
// Currently TLS can only be enabled/disabled but cannot be configured with any
// custom TLS certificates. In practice this means the default OS root certificates
// are used which should suffice for most use cases. It does mean any self-signed
// certificate is rejected.
type TLSConfig struct {
	Enable bool `yaml:"enable"`
}

// NewGrpcConnection establishes a connection with a grpc service.
func NewGrpcConnection(cfg *ServiceConfig) *grpc.ClientConn {
	opts := make([]grpc.DialOption, 0)

	if cfg.TLS != nil && cfg.TLS.Enable {
		// An empty TLS configuration defaults to the OS root certificates
		creds := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Connect to service
	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		panic(fmt.Errorf("connecting to grpc service %s: %w", cfg.Address, err))
	}

	return conn
}
