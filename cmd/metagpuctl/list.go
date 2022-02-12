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
	to.header = table.Row{"Device UUID", "Pid", "GpuMemory", "Command", "Pod", "Namespace", "Metagpus"}

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
				to.body, to.footer = composeProcessListAndFooter(resp.DevicesProcesses)
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
		to.body, to.footer = composeProcessListAndFooter(resp.DevicesProcesses)
		to.buildTable()
		to.print()
	}
}

func composeProcessListAndFooter(devProc []*pbdevice.DeviceProcess) (body []table.Row, footer table.Row) {
	var totalRequest int64
	var totalMemory uint64
	for _, deviceProcess := range devProc {
		totalRequest += deviceProcess.MetagpuRequests
		totalMemory += deviceProcess.Memory / (1024 * 1024)
		body = append(body, table.Row{
			deviceProcess.Uuid,
			deviceProcess.Pid,
			deviceProcess.Memory / (1024 * 1024),
			deviceProcess.Cmdline,
			deviceProcess.PodName,
			deviceProcess.PodNamespace,
			deviceProcess.MetagpuRequests,
		})
	}
	footer = table.Row{"Totals:", "", fmt.Sprintf("%dMb", totalMemory), "", len(devProc), "", totalRequest}
	return
}
