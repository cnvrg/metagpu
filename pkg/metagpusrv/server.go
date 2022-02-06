package metagpusrv

import (
	"fmt"
	devicevpb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	devicevapi "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/metagpusrv/deviceapi/device/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
)

func StartMetaGpuServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("api.grpc.address"), viper.GetString("metagpu-server-addr")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Infof("grpc server listening on %s:%s", viper.GetString("api.grpc.address"), viper.GetString("metagpu-server-addr"))

	grpcServer := grpc.NewServer()

	dp := devicevapi.DeviceService{}
	devicevpb.RegisterDeviceServiceServer(grpcServer, &dp)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
