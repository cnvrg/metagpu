package main

import (
	"context"
	"fmt"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/atomicgo/cursor"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
)

type TableOutput struct {
	data         []byte
	header       table.Row
	footer       table.Row
	body         []table.Row
	lastPosition int
}

func (o *TableOutput) rowsCount() int {
	return 2 + len(o.body)
}

func (o *TableOutput) Write(data []byte) (n int, err error) {
	o.data = append(o.data, data...)
	return len(data), nil
}

func (o *TableOutput) print() {
	if o.lastPosition > 0 {
		cursor.ClearLinesUp(o.lastPosition)
	}
	fmt.Printf("%s", o.data)
	o.lastPosition = o.rowsCount()
}

func (o *TableOutput) buildTable() {
	o.data = nil
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t := table.NewWriter()
	t.SetOutputMirror(o)
	t.AppendHeader(o.header, rowConfigAutoMerge)
	t.AppendRows(o.body)
	t.SetStyle(table.StyleColoredGreenWhiteOnBlack)
	t.AppendFooter(o.footer)
	t.SetColumnConfigs([]table.ColumnConfig{{Number: 1, AutoMerge: true}})
	t.SortBy([]table.SortBy{{Name: "Device UUID", Mode: table.Asc}})
	t.Render()
}

func GetGrpcMetaGpuSrvClientConn() (*grpc.ClientConn, error) {
	log.Infof("initiating gRPC connection to %s", viper.GetString("addr"))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(viper.GetString("addr"), opts...)
	if err != nil {
		return nil, err
	}
	if err := pingServer(conn); err != nil {
		log.Errorf("failed to connect to server 🙀, err: %s", err)
		os.Exit(1)
	} else {
		log.Infof("connected to %s", viper.GetString("addr"))
	}
	return conn, nil
}

func authenticatedContext() context.Context {
	ctx := context.Background()
	md := metadata.Pairs("Authorization", viper.GetString("token"))
	return metadata.NewOutgoingContext(ctx, md)
}

func getTotalRequests(processes []*pbdevice.DeviceProcess) (totalRequest int) {
	metaGpuPodRequests := make(map[string]bool)
	for _, deviceProcess := range processes {
		if _, ok := metaGpuPodRequests[deviceProcess.PodName]; !ok {
			totalRequest += int(deviceProcess.MetagpuRequests)
			metaGpuPodRequests[deviceProcess.PodName] = true
		}
	}
	return
}

func getTotalShares(devices []*pbdevice.Device) (totalShares int) {
	for _, d := range devices {
		totalShares += int(d.Shares)
	}
	return
}

func getTotalMemoryUsedByProcesses(processes []*pbdevice.DeviceProcess) (totalUsedMem int) {
	for _, p := range processes {
		totalUsedMem += int(p.Memory)
	}
	return
}

func getDeviceLoad(device *pbdevice.Device) string {
	return fmt.Sprintf("%d [GPU:%d%%|MEM:%d%%|TOT:%dMB]",
		device.Index,
		device.GpuUtilization,
		device.MemoryUsed*100/device.MemoryTotal,
		device.MemoryTotal)
}

func getDeviceByUuid(uuid string, devices []*pbdevice.Device) (device *pbdevice.Device) {
	for _, d := range devices {
		if uuid == d.Uuid {
			return d
		}
	}
	return
}
