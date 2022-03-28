package allocator

type DeviceLoad struct {
	Metagpus []string
}

type DeviceAllocation struct {
	LoadMap             []*DeviceLoad
	AvailableDevIds     []string
	AllocationSize      int
	TotalSharesPerGpu   int
	MetagpusAllocations []string
}
