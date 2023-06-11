# Introduction
The K8s Executor deploys other executors on Kubernetes.

# Usage
## Deploy an executor 
```json
{
    "conditions": {
        "executortype": "k8s"
    },
    "funcname": "deploy",
    "args": [
        "sleep-executor", <-- deployment name
        2, <-- number of pods 
        5, <-- executors per pods
        false, <-- enable a shared ramdisk
        "colonyos/sleepexecutor:v0.0.1" <-- executor container image
    ]
}
```
```console
colonies function submit --spec ./deploy.json
```

## List deployments
```console
colonies function exec --func list --targettype k8s --out --wait  
```
Output:
```json
["sleep-executor"]
```

## Scale
```json
{
    "conditions": {
        "executortype": "k8s"
    },
    "funcname": "scale",
    "args": [
        "sleep-executor", <-- deployment name
        3 <-- number of pods
    ]
}
```
```console
colonies function submit --spec ./scale.json
```
