package gpumgr

import (
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type GpuProcess struct {
	Pid            uint32
	DeviceUuid     string
	GpuUtilization uint32
	GpuMemory      uint64
	Cmdline        []string
	User           string
	ContainerId    string
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

func (p *GpuProcess) GetShortCmdLine() string {
	if len(p.Cmdline) == 0 {
		return "-"
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
	return p
}
