# Introduction
This repo contains implementation of several Colonies executors. 

| Executor | Purpose                                            |
| :---     | :-----------                                       |
| backup   | An executor to backup PostgreSQL databases.        |
| sleep    | An executor that sleeps. Can be used for testings. |
| k8s      | Deploys other executors on Kubernetes.             |

## Colonies functions
| Executor | Description                 | Function                                                                                     | Example JSON                                                                                                                  |
| :---     | :-----------                | :-----------                                                                                 | :-----------                                                                                                                  |
| backup   | Trigger a backup            | **backup**()                                                                                 |                                                                                                                               |
| sleep    | Sleep                       | **sleep**(milliseconds::string)                                                              |                                                                                                                               |
| k8s      | Deploy an executor on K8s   | **deploy**(name::string, pods::int, executorperpod::int, ramdisk::bool, dockerimage::string) | {"conditions":{"executortype":"k8s"},"funcname":"deploy","args":["sleep-executor",2,5,false,"colonyos/sleepexecutor:v0.0.1"]} |
| k8s      | Undeploy an executor on K8s | **undeploy**(name::string)                                                                   |                                                                                                                               |

