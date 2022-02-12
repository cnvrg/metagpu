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
	if viper.GetBool("watch") {
		request := &pbdevice.StreamDeviceProcessesRequest{PodId: hostname}
		stream, err := device.StreamDeviceProcesses(authenticatedContext(), request)
		if err != nil {
			log.Fatal(err)
		}

		refreshCh := make(chan bool, 1)
		refreshCh <- true
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			select {
			case s := <-sigCh:

				cursor.DownAndClear(6)
				log.Infof("signal: %s, shutting down", s)
				close(sigCh)
				close(refreshCh)
				cursor.Show()
				os.Exit(0)
			case <-refreshCh:
				cursor.Hide()
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("error watching gpu processes, err: %s", err)
				}
				printProcessesTable(resp.DevicesProcesses)
				time.Sleep(1 * time.Second)
				refreshCh <- true

			}

		}
	} else {
		ldr := &pbdevice.ListDeviceProcessesRequest{PodId: hostname}
		resp, err := device.ListDeviceProcesses(authenticatedContext(), ldr)
		if err != nil {
			log.Errorf("falid to list device processes, err: %s ", err)
			return
		}
		printProcessesTable(resp.DevicesProcesses)
	}

}

type TableOutput struct {
	data []byte
	rows int
}

func (t *TableOutput) Write(data []byte) (n int, err error) {
	t.data = append(t.data, data...)
	return len(data), nil
}

func (t *TableOutput) print() {
	fmt.Printf("%s", t.data)
	if viper.GetBool("watch") {
		cursor.StartOfLineUp(t.rows)
	}
}

func printProcessesTable(processes []*pbdevice.DeviceProcess) {
	t := table.NewWriter()
	to := &TableOutput{rows: 2}
	t.SetOutputMirror(to)
	header := table.Row{
		"Device UUID",
		"Pid",
		"GpuMemory",
		"Command",
		"Pod",
		"Namespace",
		"Metagpus",
	}
	t.AppendHeader(header)
	var rows []table.Row
	for _, deviceProcess := range processes {
		to.rows++
		rows = append(rows, table.Row{
			deviceProcess.Uuid,
			deviceProcess.Pid,
			deviceProcess.Memory / (1024 * 1024),
			deviceProcess.Cmdline,
			deviceProcess.PodName,
			deviceProcess.PodNamespace,
			deviceProcess.MetagpuRequests,
		})
	}
	t.AppendRows(rows)
	t.SetStyle(table.StyleColoredGreenWhiteOnBlack)
	t.AppendFooter(table.Row{"", "", "Free: 57%", "", "", "", "Total: 8"})
	t.Render()
	to.print()
}
