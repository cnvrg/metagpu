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
	for deviceName, processesList := range resp.Processes {
		log.Infof("Device UUID: %s", deviceName)
		for _, process := range processesList.DeviceProcess {
			log.Info(process)
		}
	}
}
