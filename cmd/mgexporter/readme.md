# Metagpu metrics exporter

### Device level metrics

Each device metric includes the following labels:

1. device id
2. device uuid

Device metrics:

* `metagpu_device_memory_total` total gpu memory for single gpu unit
* `metagpu_device_memory_free` free gpu memory per single gpu unit
* `metagpu_device_memory_used` total memory used per single gpu unit
* `metagpu_device_shares` total amount of shares per single gpu unit
* `metagpu_device_memory_share_size` amount of memory for each gpu share

Metrics example: 
```
metagpu_device_memory_free{device_index="0",device_uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 8790
metagpu_device_memory_share_size{device_index="0",device_uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 1040
metagpu_device_memory_total{device_index="0",device_uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 11441
metagpu_device_memory_used{device_index="0",device_uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 2650
metagpu_device_shares{device_index="0",device_uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 11
```

### Process level metrics 

Each process metrics includes the following labels:
1. process pid
2. cmdline
3. pod name
4. pod namespace
5. user 
6. device uuid 

Process metrics:
* `metagpu_process_gpu_utilization` process gpu utilization - calculated from device level totals

* `metagpu_process_memory_usage` process memory usage

* `metagpu_process_metagpu_requests` total quantity of metagpu requests

* `metagpu_process_max_allowed_metagpu_gpu_utilization` 
max allowed metagpu GPU utilization, calculated by: 
`metagpu_process_metagpu_requests` * `metagpu_device_memory_share_size`

* `metagpu_process_max_allowed_metagpu_memory` max allowed metagpu memory usage
calculated by: `metagpu_process_metagpu_requests` * `metagpu_device_memory_share_size` 
 
* `metagpu_process_metagpu_current_gpu_utilization` current gpu utilization
calculated by: `metagpu_process_gpu_utilization` * 100 / `metagpu_process_max_allowed_metagpu_memory`
 
* `metagpu_process_metagpu_current_memory_utilization` current memory utilization calculated by:
`metagpu_process_memory_usage` * 100 / `metagpu_process_max_allowed_metagpu_memory`

Metrics example:
```bash
metagpu_process_gpu_utilization{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 44
metagpu_process_max_allowed_metagpu_gpu_utilization{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 36
metagpu_process_max_allowed_metagpu_memory{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 4160
metagpu_process_memory_usage{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 2643
metagpu_process_metagpu_current_gpu_utilization{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 122
metagpu_process_metagpu_current_memory_utilization{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 63
metagpu_process_metagpu_requests{cmdline="/usr/local/bin/python",pid="3954530",pod_name="gpu-test-with-gpu-74f754c674-99gdb",pod_namespace="default",user="root",uuid="GPU-92fbf3b0-28f0-1add-7cd7-255fbdbd6e53"} 4
```
