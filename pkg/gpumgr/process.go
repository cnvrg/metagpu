package gpumgr

import (
	"context"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/sharecfg"
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1core "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
)

type GpuProcess struct {
	Pid               uint32
	DeviceUuid        string
	GpuUtilization    uint32
	GpuMemory         uint64
	Cmdline           []string
	User              string
	ContainerId       string
	PodId             string
	PodNamespace      string
	PodMetagpuRequest int64
	ResourceName      string
	Nodename          string
}

func (p *GpuProcess) SetProcessCmdline() {
	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		var e error
		p.Cmdline, e = pr.CmdlineSlice()
		if e != nil {
			log.Error(e)
		}
	} else {
		log.Error(err)
	}
}

func (p *GpuProcess) SetProcessUsername() {
	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		var e error
		p.User, e = pr.Username()
		if e != nil {
			log.Error(e)
		}
	} else {
		log.Error(err)
	}
}

func (p *GpuProcess) Kill() error {
	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		return pr.Kill()
	} else {
		return err
	}
}

func (p *GpuProcess) SetProcessContainerId() {
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
	ExitContainerIdSet:
		if p.ContainerId == "" {
			for _, g := range cgroups {
				for _, c := range g.Controllers {
					if c == "memory" {
						p.ContainerId = filepath.Base(g.Path)
						goto ExitContainerIdSet
					}
				}
			}
			log.Warnf("unable to set containerId for pid: %d", p.Pid)
		}
	}
}

func (p *GpuProcess) EnrichProcessK8sInfo() {
	c, err := podexec.GetK8sClient()
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
				log.Error("can't detect container ID form k8s pod")
				return
			}
			if cId[1] == p.ContainerId {
				p.PodId = pod.Name
				p.PodNamespace = pod.Namespace
				for _, container := range pod.Spec.Containers {
					resourceName := v1core.ResourceName(p.ResourceName)
					if quantity, ok := container.Resources.Limits[resourceName]; ok {
						p.PodMetagpuRequest = quantity.Value()
						p.Nodename = pod.Spec.NodeName
					}
				}
			}
		}
	}
}

func (p *GpuProcess) GetShortCmdLine() string {
	if len(p.Cmdline) == 0 {
		return "error discovering process cmdline"
	}
	return p.Cmdline[0]
}

func (p *GpuProcess) GetDevice(devices []*GpuDevice) *GpuDevice {
	for _, device := range devices {
		if device.UUID == p.DeviceUuid {
			return device
		}
	}
	return nil
}

func (p *GpuProcess) SetResourceName() {
	for _, cfg := range sharecfg.NewDeviceSharingConfig().Configs {
		for _, uuid := range cfg.Uuid {
			if p.DeviceUuid == uuid || uuid == "*" {
				p.ResourceName = cfg.ResourceName
				return
			}
		}
	}
}

func NewGpuProcess(pid, gpuUtil uint32, gpuMem uint64, devUuid string) *GpuProcess {
	p := &GpuProcess{
		Pid:            pid,
		GpuUtilization: gpuUtil,
		GpuMemory:      gpuMem,
		DeviceUuid:     devUuid,
	}
	p.SetProcessUsername()
	p.SetProcessCmdline()
	p.SetProcessContainerId()
	p.SetResourceName()
	p.EnrichProcessK8sInfo()
	if viper.GetBool("mgctlAutoInject") {
		podexec.CopymgctlToContainer(p.ContainerId)
	}
	return p
}

func NewGpuPod(podId, ns, resourceName, nodename string, metagpuRequests int64) *GpuProcess {
	p := &GpuProcess{
		PodId:             podId,
		PodNamespace:      ns,
		PodMetagpuRequest: metagpuRequests,
		ResourceName:      resourceName,
		Nodename:          nodename,
	}
	if viper.GetBool("mgctlAutoInject") {
		podexec.CopymgctlToContainer(p.ContainerId)
	}
	return p
}
