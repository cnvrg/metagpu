package allocator

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAllocator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Allocation Suite")
}

var _ = Describe("Metagpu Allocation testing", func() {

	Context("allocate", func() {
		It("10% gpu", func() {
			physDevs := 2
			allocationSize := 1
			sharesPerGpu := 10
			testDevices := getTestDevicesIds(physDevs, sharesPerGpu)
			alloc := NewDeviceAllocation(physDevs, allocationSize, sharesPerGpu, testDevices)
			Expect(len(alloc.MetagpusAllocations)).To(Equal(1))
			expectedDevices := []string{"cnvrg-meta-0-0-test-device-0"}
			Expect(alloc.MetagpusAllocations).To(Equal(expectedDevices))
		})
		It("50% gpu", func() {
			physDevs := 2
			sharesPerGpu := 10
			allocationSize := 5
			testDevices := getTestDevicesIds(physDevs, sharesPerGpu)
			alloc := NewDeviceAllocation(physDevs, allocationSize, sharesPerGpu, testDevices)
			Expect(len(alloc.MetagpusAllocations)).To(Equal(5))
			Expect(alloc.MetagpusAllocations).To(Equal(getTestDevicesIds(1, 5)))
		})
		It("80% gpu", func() {
			physDevs := 2
			sharesPerGpu := 10
			allocationSize := 8
			testDevices := getTestDevicesIds(physDevs, sharesPerGpu)
			alloc := NewDeviceAllocation(physDevs, allocationSize, sharesPerGpu, testDevices)
			Expect(len(alloc.MetagpusAllocations)).To(Equal(8))
			Expect(alloc.MetagpusAllocations).To(Equal(getTestDevicesIds(1, 8)))
		})
		It("100% gpu", func() {
			physDevs := 2
			allocationSize := 10
			sharesPerGpu := 10
			testDevices := getTestDevicesIds(physDevs, sharesPerGpu)
			alloc := NewDeviceAllocation(physDevs, allocationSize, sharesPerGpu, testDevices)
			Expect(len(alloc.MetagpusAllocations)).To(Equal(10))
			Expect(alloc.MetagpusAllocations).To(Equal(getTestDevicesIds(1, 10)))
		})

		It("110% gpu", func() {
			physDevs := 2
			allocationSize := 12
			sharesPerGpu := 10
			testDevices := getTestDevicesIds(physDevs, sharesPerGpu)
			alloc := NewDeviceAllocation(physDevs, allocationSize, sharesPerGpu, testDevices)
			Expect(len(alloc.MetagpusAllocations)).To(Equal(12))
			expectedIds := getTestDevicesIds(1, 10)
			expectedIds = append(expectedIds, "cnvrg-meta-1-0-test-device-1")
			expectedIds = append(expectedIds, "cnvrg-meta-1-1-test-device-1")
			Expect(alloc.MetagpusAllocations).To(Equal(expectedIds))
		})

		It("200% gpu", func() {
			physDevs := 2
			allocationSize := 20
			sharesPerGpu := 10
			testDevices := getTestDevicesIds(physDevs, sharesPerGpu)
			alloc := NewDeviceAllocation(physDevs, allocationSize, sharesPerGpu, testDevices)
			Expect(len(alloc.MetagpusAllocations)).To(Equal(20))
			Expect(alloc.MetagpusAllocations).To(Equal(getTestDevicesIds(2, 10)))
		})

	})
})

func getTestDevicesIds(physicalDevices, sharesPerGpu int) (metagpus []string) {
	for i := 0; i < physicalDevices; i++ {
		for j := 0; j < sharesPerGpu; j++ {
			metagpus = append(metagpus, fmt.Sprintf("cnvrg-meta-%d-%d-test-device-%d", i, j, i))
		}
	}
	return
}
