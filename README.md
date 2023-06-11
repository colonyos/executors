# Introduction
This repo contains implementation of several Colonies executors. 

## Executors 
| Executor | Purpose                                            |
| :---     | :-----------                                       |
| backup   | An executor to backup PostgreSQL databases.        |
| sleep    | An executor that sleeps. Can be used for testings. |
| k8s      | Deploys other executors on Kubernetes.             |

## Colonies functions
### Backup executor
| Function     | Description      |
| :----------- | :-----------     |
| **backup**() | Trigger a backup |

### Sleep executor
| Function                        | Description  |
| :-----------                    | :----------- |
| **sleep**(milliseconds::string) | Sleep.       |

### K8s executor
| Function                                                                                               | Description                       |
| :-----------                                                                                           | :-----------                      |
| **deploy**(deploymentname::string, pods::int, executorperpod::int, ramdisk::bool, dockerimage::string) | Deploy an executor on K8s.        |
| **undeploy**(deploymentname::string)                                                                   | Undeploy an executor on K8s       |
| **get_deployments**()                                                                                  | List deployments                  | 
| **scale**(deploymentname::string, pods::int)                                                           | Scale deployments                 |
| **get_scale**(deploymentname::string)                                                                  | Get scale factor of a deployments |
| **get_pods**(deploymentname::string)                                                                   | List pods                         |
| **pods**(deploymentname::string)                                                                       | Get number of pods                |
| **get_containers**(podname::string)                                                                    | List container names              |
| **containers**(podname::string)                                                                        | Get number of containers          |
| **restart**(podname::string)                                                                           | Restart pod                       |
