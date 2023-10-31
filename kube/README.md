# Introduction
TODO This executor runs Unix commands. 

![Design](docs/KubeExecutorDesign.png)

## Usage

```json
{
    "conditions": {
        "executortype": "kubeexecutor",
        "nodes": 2,
        "processes-per-node": 2 
        "mem": "1Gi",
        "cpu": "500m",
        "walltime": 60,
        "gpu": {
            "count": 0
        }
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "echo",
        "docker-image": "busybox",
        "rebuild-image": false,
        "args": [
            "Hello, World!"
        ],
        "keep_snapshots": false
    },
    "fs": {
        "mount": "/cfs",
        "snapshots": []
    },
    "maxwaittime": -1,
    "maxexectime": 100,
    "maxretries": 3
}
```
