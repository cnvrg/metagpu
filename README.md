# MetaGPU Device Plugin for Kubernetes

The metagpu device plugin (`mgdp`) allows you to share one or more Nvidia GPUs between
different K8s workloads. 

### Motivation
K8s doesn't provide a support for the GPU sharing. 
Meaning user must allocate entire GPU to his workload, even if the actual GPU usage 
is much bellow of 100%. 
This project will help to improve the GPU utilization by allowing GPU sharing between 
multiple K8s workloads. 


### How it works 
The `mgdp` is based on [Nvidia Container Runtime](https://github.com/NVIDIA/nvidia-container-runtime)
and on [go-nvml](https://github.com/NVIDIA/go-nvml)
One for the features the nvidia container runtime providers, is an ability 
to specify the visible GPU devices Ids by using env vars `NVIDIA_VISIBLE_DEVICES`.


### [MetaGPU demo from Cnvrg's MLCon 2.0](https://www.youtube.com/watch?v=hsP9GXUtNNs)

The most short & simple explanation of the `mgdp` logic is:
1. `mgdp` detects all the GPU devices Ids 
2. From the real GPU deices Ids, it's generates a meta-devices Ids
3. `mgdp` advertise these meta-devices Ids to the K8s
4. Once a user requests for a gpu fraction, for example 0.5 GPU, `mgdp` will allocate 50 meta-devices IDs
5. The 50 meta-gpus are bounded to 1 real device id, this real device ID will be injected to the container 

In addition, each metagpu container will have `mgctl` binary. 
The `mgctl` is an alternative for `nvidia-smi`. 
The `mgctl` improves security and provides better K8s integration.

### The sharing configurations
By default, `mgdp` will share each of your GPU devices to 100 meta-gpus. 
For example, if you've a machine with 2 GPUs, `mgdp` will generate 200 metagpus. 
Requesting for 50 metagpus, will give you 0.5 GPU, requesting 150 metagpus, 
will give you 1.5 metagpus.


 



