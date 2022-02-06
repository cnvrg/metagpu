package metagpusrv

import (
	"context"
	"fmt"
	devicevpb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/deviceplugin"
	devicevapi "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/metagpusrv/deviceapi/device/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"time"
)

type MetaGpuServer struct {
	plugin *deviceplugin.MetaGpuDevicePlugin
}

func NewMetaGpuServer(plugin *deviceplugin.MetaGpuDevicePlugin) *MetaGpuServer {
	return &MetaGpuServer{plugin: plugin}
}

func (s *MetaGpuServer) Start() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s", viper.GetString("metagpu-server-addr")))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Infof("grpc server listening on %s", viper.GetString("metagpu-server-addr"))

		opts := []grpc.ServerOption{grpc.UnaryInterceptor(s.unaryServerInterceptor())}

		grpcServer := grpc.NewServer(opts...)

		dp := devicevapi.DeviceService{}
		devicevpb.RegisterDeviceServiceServer(grpcServer, &dp)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *MetaGpuServer) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		ctx = context.WithValue(ctx, "plugin", s.plugin)
		h, err := handler(ctx, req)
		log.Infof("[method: %s duration: %s]", info.FullMethod, time.Since(start))
		return h, err
	}
}
