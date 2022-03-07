package utils

import (
	"context"
	"fmt"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/spf13/viper"
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

func NvmlErrorCheck(ret nvml.Return) {
	if ret == nvml.ERROR_NOT_FOUND {
		log.Warnf("nvml error: ERROR_NOT_FOUND: [a query to find an object was unsuccessful]")
		return
	}
	if ret == nvml.ERROR_NOT_SUPPORTED {
		log.Warnf("nvml error: ERROR_NOT_SUPPORTED: [device doesn't support this feature]")
		return
	}
	if ret == nvml.ERROR_NO_PERMISSION {
		log.Warnf("nvml error: ERROR_NO_PERMISSION: [user doesn't have permission to perform this operation]")
		return
	}
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}

func getGpuSharesByResourceName(devUuid string) (int, error) {
	type deviceConfig struct {
		uuid         string
		resourceName string
		metaGpus     int
	}
	var devConfig []deviceConfig
	if err := viper.UnmarshalKey("deviceIds", devConfig); err != nil {
		log.Error(err, "bad configs")
		os.Exit(1)
	}
	for _, c := range devConfig {
		if c.uuid == devUuid {
			return c.metaGpus, nil
		}
	}
	return -1, fmt.Errorf("gpuShares not found for %s device", devUuid)

}
