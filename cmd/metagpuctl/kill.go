package main

import (
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill process",
	Run: func(cmd *cobra.Command, args []string) {
		killGpuProcess()
	},
}

func killGpuProcess() {
	conn, err := GetGrpcMetaGpuSrvClientConn()
	if err != nil {
		log.Fatalf("can't initiate connection to metagpu server, %s", err)
	}
	device := pbdevice.NewDeviceServiceClient(conn)
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("faild to detect podId, err: %s", err)
	}
	ldr := &pbdevice.ListDeviceProcessesRequest{PodId: hostname}
	resp, err := device.ListDeviceProcesses(authenticatedContext(), ldr)
	if err != nil {
		log.Errorf("falid to list device processes, err: %s ", err)
		return
	}

	killProcessTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   `> {{ printf "[Pid:%d] %s" .Pid .Uuid | cyan }}`,
		Inactive: `  {{ printf "[Pid:%d] %s" .Pid .Uuid | faint }}`,
		Selected: `> {{ printf "[Pid:%d] %s" .Pid .Uuid | cyan }}`,
		Details: `
--------- Kill GPU process  ----------
{{ "Cmd:" | faint }}	{{ .Cmdline }}
{{ "GpuMemory:" | faint }}	{{ .Memory }}
{{ "Pod name:" | faint }}	{{ .PodName }}
{{ "Pod namespace:" | faint }}	{{ .PodNamespace }}`,
	}

	killProcessSelect := promptui.Select{
		Label:     "Select a backup",
		Items:     resp.DevicesProcesses,
		Size:      10,
		Templates: killProcessTemplate,
	}
	idx, _, err := killProcessSelect.Run()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(idx)
}
