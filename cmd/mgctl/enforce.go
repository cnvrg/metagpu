package main

import (
	"github.com/spf13/cobra"
)

var enforceCmd = &cobra.Command{
	Use:   "enforce",
	Short: "enforce memory limits",
	Run: func(cmd *cobra.Command, args []string) {
		//enforceMemoryLimits()
	},
}

////
////func enforceMemoryLimits() {
////	conn := ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString(flagAddr))
////	if conn == nil {
////		log.Fatalf("can't initiate connection to metagpu server")
////	}
////	defer conn.Close()
////	device := pbdevice.NewDeviceServiceClient(conn)
////	hostname, err := os.Hostname()
////	if err != nil {
////		log.Errorf("faild to detect podId, err: %s", err)
////	}
////	request := &pbdevice.StreamProcessesRequest{PodId: hostname}
////	stream, err := device.StreamProcesses(ctlutils.AuthenticatedContext(viper.GetString("token")), request)
////	if err != nil {
////		log.Fatal(err)
////	}
////
////	refreshCh := make(chan bool)
////	sigCh := make(chan os.Signal, 1)
////	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
////
////	to := &TableOutput{}
////	to.header = table.Row{"Idx", "Pod", "Used Mem", "Meta Mem"}
////
////	go func() {
////		for {
////			time.Sleep(1 * time.Second)
////			refreshCh <- true
////		}
////	}()
////
////	for {
////		select {
////		case <-sigCh:
////			cursor.ClearLine()
////			log.Info("shutting down")
////			os.Exit(0)
////		case <-refreshCh:
////			processResp, err := stream.Recv()
////			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				log.Fatalf("error watching gpu processes, err: %s", err)
//			}
//			deviceResp, err := device.GetDevices(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetDevicesRequest{})
//			if err != nil {
//				log.Errorf("falid to list devices, err: %s ", err)
//				return
//			}
//			to.body, to.footer = composeMemEnforceListAndFooter(processResp.DevicesProcesses, deviceResp.Device)
//			to.buildTable()
//			to.print()
//
//			for _, p := range processResp.DevicesProcesses {
//				d := deviceResp.Device[p.Uuid]
//				if p.Memory > d.MemoryShareSize*uint64(p.MetagpuRequests) {
//					killRequest := &pbdevice.KillGpuProcessRequest{Pid: p.Pid}
//					_, _ = device.KillGpuProcess(ctlutils.AuthenticatedContext(viper.GetString("token")), killRequest)
//				}
//			}
//		}
//	}
//}
//
//func composeMemEnforceListAndFooter(processes []*pbdevice.DeviceProcess, devices map[string]*pbdevice.Device) (body []table.Row, footer table.Row) {
//
//	type enforceObj struct {
//		uuid    string
//		podName string
//		memUsed uint64
//		maxMem  uint64
//	}
//
//	var el = make(map[string]*enforceObj)
//
//	for _, p := range processes {
//		d := devices[p.Uuid]
//		el[p.PodName] = &enforceObj{
//			uuid:    p.Uuid,
//			podName: p.PodName,
//			memUsed: p.Memory,
//			maxMem:  d.MemoryShareSize * uint64(p.MetagpuRequests),
//		}
//	}
//
//	for _, eObj := range el {
//		podName := fmt.Sprintf("\033[32m%s\033[0m", eObj.podName)
//		if eObj.memUsed > eObj.maxMem {
//			podName = fmt.Sprintf("\033[31m%s\033[0m", eObj.podName)
//		}
//		body = append(body, table.Row{eObj.uuid, podName, eObj.memUsed, eObj.maxMem})
//	}
//
//	footer = table.Row{"", "", "", ""}
//
//	return
//
//}
