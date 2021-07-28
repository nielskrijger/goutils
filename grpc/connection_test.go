package grpc_test

import (
	"testing"

	"github.com/nielskrijger/goutils/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/connectivity"
)

func TestNewGrpcConnection_Success(t *testing.T) {
	conn := grpc.NewGrpcConnection(&grpc.ServiceConfig{
		Address: "test:50051",
	})

	assert.Equal(t, "test:50051", conn.Target())
	assert.Equal(t, connectivity.Idle, conn.GetState())
}

func TestNewGrpcConnection_WithTLSSuccess(t *testing.T) {
	conn := grpc.NewGrpcConnection(&grpc.ServiceConfig{
		Address: "test:50051",
		TLS: &grpc.TLSConfig{
			Enable: true,
		},
	})

	assert.Equal(t, "test:50051", conn.Target())
	assert.Equal(t, connectivity.Idle, conn.GetState())
}
