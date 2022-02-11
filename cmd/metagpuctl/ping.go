package main

import (
	"context"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var PingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping server to check connectivity",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = GetGrpcMetaGpuSrvClientConn()
	},
}

func pingServer(conn *grpc.ClientConn) error {
	deviceSvc := pbdevice.NewDeviceServiceClient(conn)
	_, err := deviceSvc.PingServer(context.Background(), &pbdevice.PingServerRequest{})
	if err != nil {
		return err
	} else {
		return nil
	}

}
