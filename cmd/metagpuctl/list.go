package main

import (
	"fmt"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/atomicgo/cursor"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list resources",
}

var processListParams = []param{
	{name: "watch", shorthand: "w", value: false, usage: "watch for the changes"},
}

var processesListCmd = &cobra.Command{
	Use:   "process",
	Short: "list gpu processes, and processes metadata",
	Run: func(cmd *cobra.Command, args []string) {
		listDevicesProcesses()
	},
}

func listDevicesProcesses() {

	conn, err := GetGrpcMetaGpuSrvClientConn()
	if err != nil {
		log.Fatalf("can't initiate connection to metagpu server, %s", err)
	}
	device := pbdevice.NewDeviceServiceClient(conn)
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("faild to detect podId, err: %s", err)
	}

	to := &TableOutput{}
	to.header = table.Row{"Device UUID", "Pod", "NS", "Pid", "Memory", "Cmd", "Req"}

	if viper.GetBool("watch") {
		request := &pbdevice.StreamDeviceProcessesRequest{PodId: hostname}
		stream, err := device.StreamDeviceProcesses(authenticatedContext(), request)
		if err != nil {
			log.Fatal(err)
		}

		refreshCh := make(chan bool)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

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

				to.body, to.footer = composeProcessListAndFooter(resp.DevicesProcesses, resp.VisibilityLevel)
				to.buildTable()
				to.print()
			}
		}
	} else {
		ldr := &pbdevice.ListDeviceProcessesRequest{PodId: hostname}
		resp, err := device.ListDeviceProcesses(authenticatedContext(), ldr)
		if err != nil {
			log.Errorf("falid to list device processes, err: %s ", err)
			return
		}
		to.body, to.footer = composeProcessListAndFooter(resp.DevicesProcesses, resp.VisibilityLevel)
		to.buildTable()
		to.print()
	}
}

func composeProcessListAndFooter(devProc []*pbdevice.DeviceProcess, vl string) (body []table.Row, footer table.Row) {
	var totalRequest int64
	var totalMemory uint64
	var totalShares int32

	devu := make(map[string]string)
	devProcTotMemPerGPU := make(map[string]uint64)
	metaGpuPodRequests := make(map[string]bool)
	for _, deviceProcess := range devProc {
		if _, ok := metaGpuPodRequests[deviceProcess.PodName]; !ok {
			totalRequest += deviceProcess.MetagpuRequests
			metaGpuPodRequests[deviceProcess.PodName] = true
		}

	}
	// fuck 1
	for _, deviceProcess := range devProc {
		devProcTotMemPerGPU[deviceProcess.Uuid] += deviceProcess.Memory
	}
	// fuck 2
	for _, deviceProcess := range devProc {
		if vl != "l0" {
			devu[deviceProcess.Uuid] = "-"
		} else {
			devu[deviceProcess.Uuid] = fmt.Sprintf("[GPU:%d%%|MEM:%d%%|TOT:%dMB]",
				deviceProcess.DeviceGpuUtilization,
				devProcTotMemPerGPU[deviceProcess.Uuid]*100/deviceProcess.DeviceMemoryTotal,
				deviceProcess.DeviceMemoryTotal,
			)
		}
	}
	// fuck 3
	for _, deviceProcess := range devProc {
		totalMemory += deviceProcess.Memory
		totalShares = deviceProcess.TotalShares // I know, this is doesn't make sense, but I have to hurry up, whisky is ending
		uuid := fmt.Sprintf("%s %s", deviceProcess.Uuid, devu[deviceProcess.Uuid])

		body = append(body, table.Row{
			uuid,
			deviceProcess.PodName,
			deviceProcess.PodNamespace,
			deviceProcess.Pid,
			deviceProcess.Memory,
			deviceProcess.Cmdline,
			deviceProcess.MetagpuRequests,
		})
	}

	metaGpuSummary := fmt.Sprintf("%d/%d", totalShares, totalRequest)
	if totalShares == 0 {
		metaGpuSummary = fmt.Sprintf("%d", totalRequest)
	}
	footer = table.Row{"Totals:", "", "", len(devProc), fmt.Sprintf("%dMb", totalMemory), "", metaGpuSummary}
	return
}
