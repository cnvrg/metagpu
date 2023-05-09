package main

import (
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill process",
	Run: func(cmd *cobra.Command, args []string) {
		//killGpuProcess()
	},
}

//func killGpuProcess() {
//	conn := ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString(flagAddr))
//	if conn == nil {
//		log.Fatalf("can't initiate connection to metagpu server")
//	}
//	defer conn.Close()
//	device := pbdevice.NewDeviceServiceClient(conn)
//	hostname, err := os.Hostname()
//	if err != nil {
//		log.Errorf("faild to detect podId, err: %s", err)
//	}
//	ldr := &pbdevice.GetGpuContainersRequest{PodId: hostname}
//	resp, err := device.GetGpuContainers(ctlutils.AuthenticatedContext(viper.GetString("token")), ldr)
//	if err != nil {
//		log.Errorf("falid to list device processes, err: %s ", err)
//		return
//	}
//
//	killProcessTemplate := &promptui.SelectTemplates{
//		Label:    "{{ . }}?",
//		Active:   `> {{ printf "[Pid:%d] %s" .Pid .Uuid | cyan }}`,
//		Inactive: `  {{ printf "[Pid:%d] %s" .Pid .Uuid | faint }}`,
//		Selected: `> {{ printf "[Pid:%d] %s" .Pid .Uuid | cyan }}`,
//		Details: `
//--------- Kill GPU process  ----------
//{{ "Cmd:" | faint }}	{{ .Cmdline }}
//{{ "GpuMemory:" | faint }}	{{ .Memory }}MB
//{{ "Pod name:" | faint }}	{{ .PodName }}
//{{ "Pod namespace:" | faint }}	{{ .PodNamespace }}`,
//	}
//
//	killProcessSelect := promptui.Select{
//		Label:     "Select a process",
//		Items:     resp.DevicesProcesses,
//		Size:      10,
//		Templates: killProcessTemplate,
//	}
//	idx, _, err := killProcessSelect.Run()
//	if err != nil {
//		log.Error(err)
//		return
//	}
//	process := resp.DevicesProcesses[idx]
//	var confirmTemplate = &promptui.SelectTemplates{
//		Label:    `{{ . }}?`,
//		Active:   `> {{ . | red}}`,
//		Inactive: `  {{ . | faint}} `,
//		Selected: `> {{ . | red }}`,
//	}
//	confirmDelete := promptui.Select{
//		Label:     fmt.Sprintf("Killing PID: %d on device: %s, are you sure?", process.Pid, process.Uuid),
//		Items:     []string{"No", "Yes"},
//		Templates: confirmTemplate,
//	}
//	_, confirm, err := confirmDelete.Run()
//	if err != nil {
//		log.Error(err)
//		return
//	}
//
//	if confirm == "Yes" {
//		killRequest := &pbdevice.KillGpuProcessRequest{Pid: process.Pid}
//		if _, err := device.KillGpuProcess(ctlutils.AuthenticatedContext(viper.GetString("token")), killRequest); err != nil {
//			log.Fatalf("error killing process, err: %s", err)
//		} else {
//			log.Infof("%d killed", process.Pid)
//		}
//	}
//}
