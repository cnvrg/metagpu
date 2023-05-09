package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/ctlutils"
	"github.com/atomicgo/cursor"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "get resources",
}

var processGetParams = []param{
	{name: "watch", shorthand: "w", value: false, usage: "watch for the changes"},
}

var processesGetCmd = &cobra.Command{
	Use:     "processes",
	Aliases: []string{"p", "process"},
	Short:   "list gpu processes and processes metadata",
	Run: func(cmd *cobra.Command, args []string) {
		getDevicesProcesses()
	},
}

var getDevicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"d", "device"},
	Short:   "get gpu devices",
	Run: func(cmd *cobra.Command, args []string) {
		getDevices()
	},
}

func getDevices() {
	conn := ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString(flagAddr))
	if conn == nil {
		log.Fatalf("can't initiate connection to metagpu server")
	}
	defer conn.Close()
	device := pbdevice.NewDeviceServiceClient(conn)
	resp, err := device.GetMetaDeviceInfo(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetMetaDeviceInfoRequest{})
	if err != nil {
		log.Fatal(err)
	}

	var printer devicesPrinter

	switch of := viper.GetString(flagOutput); of {
	case outTable:
		printer = &devicesPrinterTable{}
	case outJSON:
		printer = &devicesPrinterJSON{pretty: viper.GetBool(flagPrettyOut)}
	case outRaw:
		printer = &devicesPrinterRAW{}
	default:
		log.Fatalf("Output format %q is not supported", of)
	}

	if err := printer.print(resp.Devices); err != nil {
		log.Fatal(err)
	}
}

func getDevicesProcesses() {

	conn := ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString(flagAddr))
	if conn == nil {
		log.Fatalf("can't initiate connection to metagpu server")
	}
	defer conn.Close()
	device := pbdevice.NewDeviceServiceClient(conn)
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("faild to detect podId, err: %s", err)
	}

	var printer deviceProcessesPrinter

	switch of := viper.GetString(flagOutput); of {
	case outTable:
		printer = &deviceProcessesPrinterTable{}
	case outJSON:
		printer = &deviceProcessesPrinterJSON{pretty: viper.GetBool(flagPrettyOut)}
	case outRaw:
		printer = &deviceProcessesPrinterRAW{}
	default:
		log.Fatalf("Output format %q is not supported", of)
	}

	if viper.GetBool("watch") {
		request := &pbdevice.StreamGpuContainersRequest{PodId: hostname}
		stream, err := device.StreamGpuContainers(ctlutils.AuthenticatedContext(viper.GetString("token")), request)
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
				deviceResp, err := device.GetDevices(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetDevicesRequest{})
				if err != nil {
					log.Errorf("falid to list devices, err: %s ", err)
					return
				}
				if err := printer.print(processResp.GpuContainers, deviceResp.Device, processResp.VisibilityLevel); err != nil {
					log.Fatal(err)
				}
			}
		}
	} else {
		processResp, err := device.GetGpuContainers(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetGpuContainersRequest{PodId: hostname})
		if err != nil {
			log.Errorf("falid to list device processes, err: %s ", err)
			return
		}
		deviceResp, err := device.GetDevices(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetDevicesRequest{})
		if err != nil {
			log.Errorf("falid to list devices, err: %s ", err)
			return
		}
		if err := printer.print(processResp.GpuContainers, deviceResp.Device, processResp.VisibilityLevel); err != nil {
			log.Fatal(err)
		}
	}
}

// devicesPrinter interface to print GPU devices.
type devicesPrinter interface {
	print([]*pbdevice.Device) error
}

// devicesPrinterTable - printer with human-readable table output format.
type devicesPrinterTable struct {
	to *TableOutput
}

// Check interface compatibility.
var _ devicesPrinter = (*devicesPrinterTable)(nil)

func (dpt *devicesPrinterTable) print(devices []*pbdevice.Device) error {
	if dpt.to == nil {
		dpt.to = &TableOutput{}
		dpt.to.header = table.Row{"Idx", "UUID", "Memory", "Shares", "Share size"}
	}

	dpt.to.body, dpt.to.footer = dpt.buildDeviceInfoTableBody(devices)
	dpt.to.buildTable()
	dpt.to.print()

	return nil
}

// buildDeviceInfoTableBody adds rows to the resulting table.
func (dpt *devicesPrinterTable) buildDeviceInfoTableBody(devices []*pbdevice.Device) (body []table.Row, footer table.Row) {
	var totMem uint64
	var shares uint32
	for _, d := range devices {
		shares = d.Shares
		totMem += d.MemoryTotal
		body = append(body, table.Row{
			d.Index,
			d.Uuid,
			d.MemoryTotal,
			d.Shares,
			d.MemoryShareSize,
		})
	}
	footer = table.Row{len(devices), "", fmt.Sprintf("%dMB", totMem), uint32(len(devices)) * shares, ""}
	return body, footer
}

// devicesPrinterTable - printer with JSON and indented JSON output formats.
type devicesPrinterJSON struct {
	pretty bool
}

// Check interface compatibility.
var _ devicesPrinter = (*devicesPrinterJSON)(nil)

func (dpt *devicesPrinterJSON) print(devices []*pbdevice.Device) error {
	var (
		bts []byte
		err error
	)

	if dpt.pretty {
		bts, err = json.MarshalIndent(devices, "", "  ")
	} else {
		bts, err = json.Marshal(devices)
	}
	if err != nil {
		return err
	}

	println(string(bts))

	return nil
}

// devicesPrinterRAW  - printer with Go internal output format.
type devicesPrinterRAW struct{}

// Check interface compatibility.
var _ devicesPrinter = (*devicesPrinterRAW)(nil)

func (dpt *devicesPrinterRAW) print(devices []*pbdevice.Device) error {
	_, err := fmt.Printf("%+v\n", devices)

	return err
}

// deviceProcessesPrinter interface to print GPU devices processes.
type deviceProcessesPrinter interface {
	print(containers []*pbdevice.GpuContainer, devices map[string]*pbdevice.Device, vl string) error
}

// deviceProcessesPrinterTable - printer with human-readable table output format.
type deviceProcessesPrinterTable struct {
	to *TableOutput
}

// Check interface compatibility.
var _ deviceProcessesPrinter = (*deviceProcessesPrinterTable)(nil)

func (dpt *deviceProcessesPrinterTable) print(containers []*pbdevice.GpuContainer, devices map[string]*pbdevice.Device, vl string) error {
	// First call - add header.
	if dpt.to == nil {
		dpt.to = &TableOutput{}
		dpt.to.header = table.Row{"Pod", "NS", "Device", "Node", "GPU", "Memory", "Pid", "Cmd", "Req"}
	}

	dpt.to.body = dpt.buildDeviceProcessesTableBody(containers)
	dpt.to.footer = dpt.buildDeviceProcessesTableFooter(containers, devices, vl)
	dpt.to.buildTable()
	dpt.to.print()

	return nil
}

// buildDeviceProcessesTableBody - adds rows to the resulting table.
func (dpt *deviceProcessesPrinterTable) buildDeviceProcessesTableBody(containers []*pbdevice.GpuContainer) (body []table.Row) {

	for _, c := range containers {
		if len(c.ContainerDevices) > 0 {
			maxMem := int64(c.ContainerDevices[0].Device.MemoryShareSize * uint64(c.MetagpuRequests))
			if len(c.DeviceProcesses) > 0 {
				for _, p := range c.DeviceProcesses {
					relativeGpuUsage := (p.GpuUtilization * 100) / (100 / c.ContainerDevices[0].Device.Shares * uint32(c.MetagpuRequests))
					gpuUsage := fmt.Sprintf("\u001B[32m%d%%\u001B[0m", relativeGpuUsage)
					if relativeGpuUsage > 100 {
						gpuUsage = fmt.Sprintf("\u001B[31m%d%%\u001B[0m", relativeGpuUsage)
					}
					memUsage := fmt.Sprintf("\u001B[32m%d\u001B[0m/%d", p.Memory, maxMem)
					if int64(p.Memory) > maxMem {
						memUsage = fmt.Sprintf("\u001B[31m%d\u001B[0m/%d", p.Memory, maxMem)
					}
					body = append(body, table.Row{
						c.PodId,
						c.PodNamespace,
						formatContainerDeviceIndexes(c),
						c.NodeName,
						gpuUsage,
						memUsage,
						p.Pid,
						p.Cmdline,
						c.MetagpuRequests,
					})
				}
			} else {
				memUsage := fmt.Sprintf("\u001B[32m%d\u001B[0m/%d", 0, maxMem)
				body = append(body, table.Row{
					c.PodId,
					c.PodNamespace,
					formatContainerDeviceIndexes(c),
					c.NodeName,
					"-",
					memUsage,
					"-",
					"-",
					c.MetagpuRequests,
				})
			}
		} else {
			body = append(body, table.Row{
				c.PodId,
				c.PodNamespace,
				formatContainerDeviceIndexes(c),
				c.NodeName,
				"-",
				"-",
				"-",
				"-",
				c.MetagpuRequests,
			})
		}

	}

	return
}

// buildDeviceProcessesTableFooter - generates footer for the table output.
func (dpt *deviceProcessesPrinterTable) buildDeviceProcessesTableFooter(containers []*pbdevice.GpuContainer, devices map[string]*pbdevice.Device, vl string) (footer table.Row) {
	// metaGpuSummary := fmt.Sprintf("%d", getTotalRequests(containers))
	// TODO: fix this, the vl should be taken from directly form the  package
	// to problem is that package now includes the nvidia linux native stuff
	// and some package re-org is required
	//if vl == "l0" { // TODO: temporary disabled
	metaGpuSummary := fmt.Sprintf("%d/%d", getTotalShares(devices), getTotalRequests(containers))
	//}
	usedMem := fmt.Sprintf("%dMb", getTotalMemoryUsedByProcesses(containers))
	return table.Row{len(containers), "", "", "", "", usedMem, "", "", metaGpuSummary}
}

// deviceProcessesPrinterJSON - printer with JSON and indented JSON output format.
type deviceProcessesPrinterJSON struct {
	pretty bool
}

// Check interface compatibility.
var _ deviceProcessesPrinter = (*deviceProcessesPrinterJSON)(nil)

func (dpt *deviceProcessesPrinterJSON) print(containers []*pbdevice.GpuContainer, _ map[string]*pbdevice.Device, vl string) error {
	type enrichedProcess struct {
		*pbdevice.DeviceProcess
		RelativeGpuUsagePercent *uint32 `json:"relative_gpu_usage_percent,omitempty"` // Can be JSON null.
	}

	type enrichedGPUContainer struct {
		*pbdevice.GpuContainer
		DeviceProcesses []*enrichedProcess `json:"device_processes,omitempty"`
	}

	var result = make([]*enrichedGPUContainer, 0, len(containers))

	for _, c := range containers {
		enrCnt := &enrichedGPUContainer{
			GpuContainer: c,
		}

		for _, p := range c.DeviceProcesses {
			if len(c.ContainerDevices) > 0 {
				relativeGpuUsage := (p.GpuUtilization * 100) / (100 / c.ContainerDevices[0].Device.Shares * uint32(c.MetagpuRequests))

				enrCnt.DeviceProcesses = append(enrCnt.DeviceProcesses, &enrichedProcess{
					DeviceProcess:           p,
					RelativeGpuUsagePercent: &relativeGpuUsage,
				})
			} else {
				enrCnt.DeviceProcesses = append(enrCnt.DeviceProcesses, &enrichedProcess{
					DeviceProcess:           p,
					RelativeGpuUsagePercent: nil,
				})
			}
		}

		result = append(result, enrCnt)
	}

	var (
		bts []byte
		err error
	)

	if dpt.pretty {
		bts, err = json.MarshalIndent(result, "", "  ")
	} else {
		bts, err = json.Marshal(result)
	}
	if err != nil {
		return err
	}

	println(string(bts))

	return nil
}

// deviceProcessesPrinterRAW - printer with JSON and indented JSON output format.
type deviceProcessesPrinterRAW struct{}

// Check interface compatibility.
var _ deviceProcessesPrinter = (*deviceProcessesPrinterRAW)(nil)

func (dpt *deviceProcessesPrinterRAW) print(containers []*pbdevice.GpuContainer, _ map[string]*pbdevice.Device, vl string) error {
	_, err := fmt.Printf("%+v\n", containers)
	return err
}
