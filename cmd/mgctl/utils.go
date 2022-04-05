package main

import (
	"fmt"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/atomicgo/cursor"
	"github.com/jedib0t/go-pretty/v6/table"
	"strings"
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
	t.Render()
}

func getTotalRequests(containers []*pbdevice.GpuContainer) (totalRequest int) {
	for _, c := range containers {
		totalRequest += int(c.MetagpuRequests)
	}
	return
}

func getTotalShares(devices map[string]*pbdevice.Device) (totalShares int) {
	for _, d := range devices {
		totalShares += int(d.Shares)
	}
	return
}

func getTotalMemoryUsedByProcesses(containers []*pbdevice.GpuContainer) (totalUsedMem int) {
	for _, c := range containers {
		for _, p := range c.DeviceProcesses {
			totalUsedMem += int(p.Memory)
		}
	}
	return
}

func formatContainerDeviceIndexes(container *pbdevice.GpuContainer) string {
	var devIdxs []string
	for _, d := range container.ContainerDevices {
		devIdxs = append(devIdxs, fmt.Sprintf("%d", d.Device.Index))
	}
	if len(devIdxs) > 0 {
		return strings.Join(devIdxs, ":")
	}
	return "-"
}
