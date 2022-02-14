package deviceplugin

import (
	"context"
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1core "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
)

type DeviceProcess struct {
	Pid                        uint32
	DeviceGpuUtilization       uint32 // TODO: shouldn't be here, go back and remove me!
	DeviceGpuMemoryUtilization uint32 // TODO: shouldn't be here, go back and remove me!
	DeviceGpuMemoryTotal       uint64 // TODO: shouldn't be here, go back and remove me!
	DeviceGpuMemoryFree        uint64 // TODO: shouldn't be here, go back and remove me!
	DeviceGpuMemoryUsed        uint64 // TODO: shouldn't be here, go back and remove me!
	TotalShares                int    // TODO: shouldn't be here, go back and remove me!
	GpuMemory                  uint64
	Cmdline                    []string
	User                       string
	ContainerId                string
	PodId                      string
	PodNamespace               string
	PodMetagpuRequest          int64
	TotalDevices               int32
}

func NewDeviceProcess(pid uint32, gpuMem uint64) *DeviceProcess {
	dp := &DeviceProcess{
		Pid:       pid,
		GpuMemory: gpuMem,
	}
	dp.SetProcessUsername()
	dp.SetProcessCmdline()
	dp.SetProcessContainerId()
	dp.EnrichProcessK8sInfo()
	if viper.GetBool("mgctlAutoInject") {
		copymgctlToContainer(dp.ContainerId)
	}
	return dp
}

func (p *DeviceProcess) SetProcessCmdline() {
	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		var e error
		p.Cmdline, e = pr.CmdlineSlice()
		checkProcessDiscoveryError(e)
	} else {
		log.Error(err)
	}
}

func (p *DeviceProcess) SetProcessUsername() {
	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		var e error
		p.User, e = pr.Username()
		checkProcessDiscoveryError(e)
	} else {
		log.Error(err)
	}
}

func (p *DeviceProcess) Kill() error {
	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		return pr.Kill()
	} else {
		return err
	}
}

func (p *DeviceProcess) SetProcessContainerId() {
	if proc, err := procfs.NewProc(int(p.Pid)); err == nil {
		var e error
		var cgroups []procfs.Cgroup
		cgroups, e = proc.Cgroups()
		if e != nil {
			log.Error(e)
		}
		if len(cgroups) == 0 {
			log.Errorf("cgroups list for %d is empty", p.Pid)
		}
		p.ContainerId = filepath.Base(cgroups[0].Path)
	}
}

func (p *DeviceProcess) EnrichProcessK8sInfo() {
	c, err := GetK8sClient()
	if err != nil {
		log.Error(err)
		return
	}
	pl := &v1core.PodList{}
	if err := c.List(context.Background(), pl); err != nil {
		log.Error(err)
		return
	}
	for _, pod := range pl.Items {
		for _, cStatus := range pod.Status.ContainerStatuses {
			cId := strings.Split(cStatus.ContainerID, "//")
			if len(cId) < 2 {
				log.Error("can't detect container id form k8s pod")
				return
			}
			if cId[1] == p.ContainerId {
				p.PodId = pod.Name
				p.PodNamespace = pod.Namespace
				for _, container := range pod.Spec.Containers {
					resourceName := v1core.ResourceName(viper.GetString("resourceName"))
					if quantity, ok := container.Resources.Limits[resourceName]; ok {
						p.PodMetagpuRequest = quantity.Value()
					}
				}
			}
		}
	}
}

func (p *DeviceProcess) GetShortCmdLine() string {
	if len(p.Cmdline) == 0 {
		return "error discovering process cmdline"
	}
	return p.Cmdline[0]
}

func checkProcessDiscoveryError(e error) {
	if e != nil {
		log.Error(e)
	}
}
