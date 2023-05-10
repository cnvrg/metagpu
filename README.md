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


### [MetaGPU demo from Cnvrg's MLCon 2.0](https://www.youtube.com/watch?v=hsP9GXUtNNs)

### Deployment 
1. clone the repo 
2. use helm chart to install or dump manifest and install manually 

### Install with helm chart 
```bash
# cd into cloned directory and run
# for openshift set ocp=true  
helm install chart --set ocp=false -ncnvrg 
```

### Install with raw K8s manifests 
```bash
# cd into cloned directory and run
# for openshift set ocp=true  
helm template chart --set ocp=false -ncnvrg > meatgpu.yaml 
kubectl apply -f meatgpu.yaml 
```

### GPU Memory enforcement and overcommit

The metagpu device plugin can be deployed with memory enforcement turned on (`memoryEnforcer: true`).

In this case, if the GPU process use more memory than defined by metagpu resource limit, `mgdp` will act as an Out-Of-Memory (OOM) killer, terminating the workload.

Overcommit is the case, when you specify limits higher than requests. For generic `memory` k8s resource, it is quite common that requests used for Pod scheduling are lower, but limits (OOM line) are higher to allow bursts in memory consumption.

Kubernetes currently does not allow overcommit for custom resources (like `cnvrg.io/metagpu`), enforcing requests must be equal to limits in the container spec.
If you want to achive overcommit with metagpu, you have to use `gpu-mem-limit.<resourceName>` annotation that re-defines limits value.

Following snippet represents the case, when 33 metagpus are requested (3 pods can be scheduled on the GPU), but enforcement kill happenes after the process reaches half of GPU memory (50 metagpus).

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: metagpu-overcommit
  annotations:
    gpu-mem-limit.cnvrg.io/metagpu: "50"
spec:
  containers:
  - name: workload
    resources:
      requests:
        cnvrg.io/metagpu: "33"
      limits:
        cnvrg.io/metagpu: "33"
```

### Test the Metagpu 
```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: metagpu-test
  namespace: cnvrg
spec:
  tolerations:
   - operator: "Exists"
  containers:
  - name: gpu-test-with-gpu
    image: tensorflow/tensorflow:latest-gpu
    command:
      - /usr/local/bin/python
      - -c
      - |
        import tensorflow as tf
        tf.get_logger().setLevel('INFO')
        gpus = tf.config.list_physical_devices('GPU')
        if gpus:
          # Restrict TensorFlow to only allocate 1GB of memory on the first GPU
          try:
            tf.config.set_logical_device_configuration(gpus[0],[tf.config.LogicalDeviceConfiguration(memory_limit=1024)])
            logical_gpus = tf.config.list_logical_devices('GPU')
            print(len(gpus), "Physical GPUs,", len(logical_gpus), "Logical GPUs")
          except RuntimeError as e:
            # Virtual devices must be set before GPUs have been initialized
            print(e)
        print("Num GPUs Available: ", len(tf.config.list_physical_devices('GPU')))
        while True:
          print(tf.reduce_sum(tf.random.normal([1000, 1000])))
    resources:
      limits:
        cnvrg.io/metagpu: "30"
EOF
```

 



