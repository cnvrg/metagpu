package main

import (
	"context"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	log "github.com/sirupsen/logrus"
)

func listDevicesProcesses() {
	conn, err := GetGrpcMetaGpuSrvClientConn()
	if err != nil {
		log.Fatalf("can't initiate connection to metagpu server, %s", err)
	}
	device := pbdevice.NewDeviceServiceClient(conn)
	ldr := &pbdevice.ListDeviceProcessesRequest{}
	resp, err := device.ListDeviceProcesses(context.Background(), ldr)
	if err != nil {
		log.Errorf("falid to list device processes, err: %s ", err)
		return
	}
	for _, deviceProcess := range resp.DevicesProcesses {
		log.Infof("Device UUID      : %s", deviceProcess.Uuid)
		log.Infof("Pid              : %d", deviceProcess.Pid)
		log.Infof("GpuMemory        : %d", deviceProcess.Memory/(1024*1024))
		log.Infof("Command          : %s", deviceProcess.Cmdline)
		log.Infof("ContainerID      : %s", deviceProcess.ContainerId)
		log.Infof("PodName          : %s", deviceProcess.ContainerId)
		log.Infof("PodNamespace     : %s", deviceProcess.ContainerId)
		log.Infof("MetagpuRequests  : %s", deviceProcess.ContainerId)

	}
}
