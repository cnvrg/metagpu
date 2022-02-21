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
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "list resources",
}

var processListParams = []param{
	{name: "watch", shorthand: "w", value: false, usage: "watch for the changes"},
}

var processesListCmd = &cobra.Command{
	Use:     "process",
	Aliases: []string{"p"},
	Short:   "list gpu processes and processes metadata",
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
	to.header = table.Row{"Idx", "Pod", "NS", "Pid", "Memory", "Cmd", "Req"}

	if viper.GetBool("watch") {
		request := &pbdevice.StreamProcessesRequest{PodId: hostname}
		stream, err := device.StreamProcesses(authenticatedContext(), request)
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
				processResp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("error watching gpu processes, err: %s", err)
				}
				deviceResp, err := device.ListDevices(authenticatedContext(), &pbdevice.ListDevicesRequest{})
				if err != nil {
					log.Errorf("falid to list devices, err: %s ", err)
					return
				}

				to.body = buildTableBody(processResp.DevicesProcesses, deviceResp.Device)
				to.footer = buildTableFooter(processResp.DevicesProcesses, deviceResp.Device, processResp.VisibilityLevel)
				to.buildTable()
				to.print()
			}
		}
	} else {
		processResp, err := device.ListProcesses(authenticatedContext(), &pbdevice.ListProcessesRequest{PodId: hostname})
		if err != nil {
			log.Errorf("falid to list device processes, err: %s ", err)
			return
		}
		deviceResp, err := device.ListDevices(authenticatedContext(), &pbdevice.ListDevicesRequest{})
		if err != nil {
			log.Errorf("falid to list devices, err: %s ", err)
			return
		}
		to.body = buildTableBody(processResp.DevicesProcesses, deviceResp.Device)
		to.footer = buildTableFooter(processResp.DevicesProcesses, deviceResp.Device, processResp.VisibilityLevel)
		to.buildTable()
		to.print()
	}
}

func buildTableBody(processes []*pbdevice.DeviceProcess, devices map[string]*pbdevice.Device) (body []table.Row) {
	for _, p := range processes {
		d := devices[p.Uuid]
		maxMem := d.MemoryShareSize * uint64(p.MetagpuRequests)
		memUsage := fmt.Sprintf("\u001B[32m%d\u001B[0m/%d", p.Memory, maxMem)
		if p.Memory > maxMem {
			memUsage = fmt.Sprintf("\u001B[31m%d\u001B[0m/%d", p.Memory, maxMem)
		}

		body = append(body, table.Row{
			getDeviceLoad(d),
			p.PodName,
			p.PodNamespace,
			p.Pid,
			memUsage,
			p.Cmdline,
			p.MetagpuRequests,
		})
	}
	return
}

func buildTableFooter(processes []*pbdevice.DeviceProcess, devices map[string]*pbdevice.Device, vl string) (footer table.Row) {
	metaGpuSummary := fmt.Sprintf("%d", getTotalRequests(processes))
	// TODO: fix this, the vl should be taken from directly form the  package
	// to problem is that package now includes the nvidia linux native stuff
	// and some package re-org is required
	if vl == "l0" {
		metaGpuSummary = fmt.Sprintf("%d/%d", getTotalShares(devices), getTotalRequests(processes))
	}
	usedMem := fmt.Sprintf("%dMb", getTotalMemoryUsedByProcesses(processes))
	return table.Row{len(devices), "", "", len(processes), usedMem, "", metaGpuSummary}
}
