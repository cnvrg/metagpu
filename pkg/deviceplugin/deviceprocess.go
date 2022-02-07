package deviceplugin

import (
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type DeviceProcess struct {
	Pid          uint32
	GpuMemory    uint64
	Cmdline      []string
	User         string
	ContainerId  string
	PodId        string
	PodNamespace string
}

func NewDeviceProcess(pid uint32, gpuMem uint64) *DeviceProcess {
	dp := &DeviceProcess{
		Pid:       pid,
		GpuMemory: gpuMem,
	}
	dp.enrichProcessInfo()
	return dp
}

func (p *DeviceProcess) enrichProcessInfo() {

	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		var e error
		p.Cmdline, e = pr.CmdlineSlice()
		checkProcessDiscoveryError(e)
		p.User, e = pr.Username()
		checkProcessDiscoveryError(e)
	} else {
		log.Error(err)
	}

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
		p.PodId, p.PodNamespace = inspectContainer(p.ContainerId)

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
