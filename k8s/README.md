# Introduction
The K8s executor deploys other executors on Kubernetes.

# Usage
## Deploy an executor 
```json
{
    "conditions": {
        "executortype": "k8s"
    },
    "funcname": "deploy",
    "args": [
        "sleep-executor",
        2,
        5,
        false,
        "colonyos/sleepexecutor:v1.0.0"
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
```json
["sleep-executor"]
```

## Get scale factor (replicas) of deployment
```console
colonies function exec --func get_scale --args sleep-executor --targettype k8s --out --wait 
```
```json
2
```

## Scale
```json
{
    "conditions": {
        "executortype": "k8s"
    },
    "funcname": "scale",
    "args": [
        "sleep-executor",
        3
    ]
}
```
```console
colonies function submit --spec ./scale.json
```

## List pods
```console
colonies function exec --func get_pods --targettype k8s --out --wait  
```
```json
["sleep-executor-59bf486c65-mzkdj","sleep-executor-59bf486c65-p5kxr"]
```

## Get numbers of pods
```console
colonies function exec --func pods --targettype k8s --out --wait  
```
```json
2
```

## List containers
```console
colonies function exec --func get_containers --args sleep-executor-77bbf9c97c-4c6sn --targettype k8s --out --wait
```
```json
["executor-0","executor-1","executor-2","executor-3","executor-4"]
```

## Get number of containers
```console
colonies function exec --func containers --args sleep-executor-77bbf9c97c-4c6sn --targettype k8s --out --wait
```
```json
5
```

## Restart a pod 
```console
colonies function exec --func restart --args sleep-executor-77bbf9c97c-4c6sn --targettype k8s
```
