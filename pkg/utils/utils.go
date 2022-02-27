package utils

import (
	"context"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
)

func GetGrpcMetaGpuSrvClientConn(address string) *grpc.ClientConn {
	log.Infof("initiating gRPC connection to %s", address)

	c, err := dial(address, 3*time.Second)
	if err != nil {
		log.Errorf("failed to connect to server 🙀, err: %s", err)
		os.Exit(1)
	}
	log.Infof("connected to %s", address)
	return c
}

func AuthenticatedContext(token string) context.Context {
	ctx := context.Background()
	md := metadata.Pairs("Authorization", token)
	return metadata.NewOutgoingContext(ctx, md)
}

func dial(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			c, e := net.DialTimeout("tcp", socket, timeout)
			if e != nil {
				log.Fatalf("error connecting to the server, e: %s", e)
			}
			return c, e
		}),
	}
	c, err := grpc.Dial(socket, opts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}
