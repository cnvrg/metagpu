package gpumgr

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"testing"
)

func TestAllocator(t *testing.T) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../config/")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("config file not found, err: %s", err)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Enforcer Suite")
}

var _ = Describe("enforcer", func() {

	Context("enforce", func() {

		It("not oom", func() {
			mgr := &GpuMgr{}
			mgr.setGpuDevices()
			if len(mgr.GpuDevices) < 0 {
				log.Fatalf("no gpu devices detected, can't continue unit testing")
			}
			mgr.gpuProcesses = []*GpuProcess{
				{
					Pid:               1000,
					GpuMemory:         mgr.GpuDevices[0].Memory.ShareSize,
					PodMetagpuRequest: 1,
					DeviceUuid:        mgr.GpuDevices[0].UUID,
				},
			}
			res := mgr.enforce()
			Expect(len(res)).To(Equal(0))
		})

		It("oom", func() {

			mgr := &GpuMgr{}
			mgr.setGpuDevices()
			if len(mgr.GpuDevices) < 0 {
				log.Fatalf("no gpu devices detected, can't continue unit testing")
			}
			mgr.gpuProcesses = []*GpuProcess{
				{
					Pid:               1000,
					GpuMemory:         mgr.GpuDevices[0].Memory.ShareSize + 1,
					PodMetagpuRequest: 1,
					DeviceUuid:        mgr.GpuDevices[0].UUID,
				},
			}
			res := mgr.enforce()
			Expect(len(res)).To(Equal(1))
		})

		It("false positive oom", func() {

			mgr := &GpuMgr{}
			mgr.setGpuDevices()
			if len(mgr.GpuDevices) < 0 {
				log.Fatalf("no gpu devices detected, can't continue unit testing")
			}
			mgr.gpuProcesses = []*GpuProcess{
				{
					Pid:               1000,
					GpuMemory:         0,
					PodMetagpuRequest: 1,
					DeviceUuid:        mgr.GpuDevices[0].UUID,
				},
			}
			res := mgr.enforce()
			Expect(len(res)).To(Equal(0))
		})
	})
})
