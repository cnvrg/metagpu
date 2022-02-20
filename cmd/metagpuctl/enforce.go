package main

import (
	"fmt"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/atomicgo/cursor"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var enforceCmd = &cobra.Command{
	Use:   "enforce",
	Short: "enforce memory limits",
	Run: func(cmd *cobra.Command, args []string) {
		enforceMemoryLimits()
	},
}

func enforceMemoryLimits() {
	conn, err := GetGrpcMetaGpuSrvClientConn()
	if err != nil {
		log.Fatalf("can't initiate connection to metagpu server, %s", err)
	}
	device := pbdevice.NewDeviceServiceClient(conn)
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("faild to detect podId, err: %s", err)
	}
	request := &pbdevice.StreamProcessesRequest{PodId: hostname}
	stream, err := device.StreamProcesses(authenticatedContext(), request)
	if err != nil {
		log.Fatal(err)
	}

	refreshCh := make(chan bool)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	to := &TableOutput{}
	to.header = table.Row{"Device UUID", "Pod", "Used Mem", "Meta Mem"}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			refreshCh <- true
		}
	}()

	for {
		select {
		case <-sigCh:
			cursor.ClearLine()
			log.Info("shutting down")
			os.Exit(0)
		case <-refreshCh:
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error watching gpu processes, err: %s", err)
			}
			to.body, to.footer = composeMemEnforceListAndFooter(resp.DevicesProcesses)
			to.buildTable()
			to.print()

			//for _, devProc := range resp.DevicesProcesses {
			//	if devProc.Memory > (devProc.DeviceMemoryTotal/(uint64(devProc.TotalShares)/uint64(devProc.TotalDevices)))*uint64(devProc.MetagpuRequests) {
			//		killRequest := &pbdevice.KillGpuProcessRequest{Pid: devProc.Pid}
			//		_, _ = device.KillGpuProcess(authenticatedContext(), killRequest)
			//	}
			//}
		}
	}
}

func composeMemEnforceListAndFooter(devProc []*pbdevice.DeviceProcess) (body []table.Row, footer table.Row) {

	type enforceObj struct {
		uuid    string
		podName string
		memUsed uint64
		metaMem uint64
	}

	var el = make(map[string]*enforceObj)

	//for _, enforceList := range devProc {
	//	el[enforceList.PodName] = &enforceObj{
	//		uuid:    enforceList.Uuid,
	//		podName: enforceList.PodName,
	//		memUsed: enforceList.Memory,
	//		metaMem: (enforceList.DeviceMemoryTotal / (uint64(enforceList.TotalShares) / uint64(enforceList.TotalDevices))) * uint64(enforceList.MetagpuRequests),
	//	}
	//}

	for _, eObj := range el {
		podName := fmt.Sprintf("\033[32m%s\033[0m", eObj.podName)
		if eObj.memUsed > eObj.metaMem {
			podName = fmt.Sprintf("\033[31m%s\033[0m", eObj.podName)
		}
		body = append(body, table.Row{eObj.uuid, podName, eObj.memUsed, eObj.metaMem})
	}

	footer = table.Row{"", "", "", ""}

	return

}
