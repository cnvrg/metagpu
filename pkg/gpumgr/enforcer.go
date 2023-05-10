package gpumgr

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func (m *GpuMgr) StartMemoryEnforcer() {
	log.Info("starting gpu memory enforcer")
	go func() {
		for {
			for _, p := range m.enforce() {
				p.Kill()
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (m *GpuMgr) enforce() (gpuProcForKill []*GpuProcess) {
	for _, c := range m.gpuContainers {
		for _, p := range c.Processes {
			if d := m.getGpuDeviceByUuid(p.DeviceUuid); d != nil {
				maxAllowedMem := d.Memory.ShareSize * uint64(c.PodMetagpuLimit)
				if p.GpuMemory > maxAllowedMem && p.Pid != 0 && maxAllowedMem > 0 {
					log.Infof("out of memory: %dMB/%dMB, pod: %s going to be terminated", p.GpuMemory, maxAllowedMem, c.PodId)
					gpuProcForKill = append(gpuProcForKill, p)
				}
			}
		}
	}
	return
}

func (m *GpuMgr) getGpuDeviceByUuid(uuid string) *GpuDevice {
	for _, d := range m.GpuDevices {
		if d.UUID == uuid {
			return d
		}
	}
	return nil
}
