# Introduction
The KubeExecutor enables execution of Colonies jobs as Kubernetes batch jobs. Notably, the same Colonies jobs can alternatively be executed on High Performance Computing (HPC) systems using the [HPCExecutor](https://github.com/colonyos/executors/tree/main/hpc) (which is based on Slurm). This flexibility enables seamless portability between Kubernetes and HPC environments.

The workflow is depicted in the following diagram:

![Design](docs/KubeExecutorDesign.png)

Here's how the KubeExecutor operates on a Kubernetes cluster:

**1. Process Assignments:** The KubeExecutor establishes a connection to the Colonies server and requests process assignments.

**2. Data Preparation:** If a filesystem is specified, the KubeExecutor downloads all necessary data to a Persistent Volume that is shared across all spawned batch jobs.

**3. Job Creation:** Subsequently, the KubeExecutor initiates the creation of a Kubernetes (K8s) batch job.

**4. Monitoring and Log Upload:** The KubeExecutor closely monitors the execution lifecycle of the batch job, ensuring that all logs are promptly uploaded to the Colonies server.

**5. Job Completion and Cleanup:**  After the batch job has complemeted, the KubeExecutor proceeds to delete the job from the Kubernetes cluster.

## Colonies function specification
Colonies jobs are described as so-called function specifications. Below is an example of a function specification.

```json
{
    "conditions": {
        "executortype": "kubeexecutor",
        "nodes": 2,
        "processes-per-node": 2 
        "mem": "1Gi",
        "cpu": "500m",
        "walltime": 90
    },
    "funcname": "execute",
    "kwargs": {
        "cmd": "echo",
        "docker-image": "busybox",
        "args": [
            "Hello, World!"
        ],
    },
    "maxexectime": 80,
    "maxretries": 3
}
```

* The **condition** object outlines the requirements for an acceptable executor. In this case, an executor must be of the **kubeexecutor** type to be assigned this function specification. 

* We specify a requirement for 2 nodes. This directly translates to the number of pods in the batch job. We also request 2 processes per node. This correlates with the number of containers within each pod. Each container is allocated 1Gi of memory and 0.5 cores.

* The **walltime** parameter determines the maximum lifespan of the Kubernetes job. If a job exceeds this limit, it's automatically removed from the Kubernetes cluster. This mechanism ensures that batch jobs that stall or hang don't persist indefinitely. In this case, the job is deleted after 90 seconds.

* The **maxexectime** parameter set the maximum time a process can run before it's automatically unassigned and sent back to the Colonies server. While **walltime** controls the lifespan of the Kubernetes job, **maxexectime** governs the lifespan of a Colonies process.
It's hence crucial that the value of **walltime** is set greater than that of **maxexectime**.

* The **maxretries** parameter sets a limit on the number of times a process can be unassigned due to issues or failures. After reaching this limit, the process is closed as failed and will no longer be eligible for reassignment to another executor.

* The **funcname** defines the name of function we would like to call on the executor. The **execute* function spawns batch jobs. Additionaly, the **kwargs** parameter represents the arguments passed to the execute function. This defines the exact process to be executed. In this particular example, we wish to run the command **echo Hello, World!** within the busybox Docker container.
