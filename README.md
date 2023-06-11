# Introduction
This repo contains implementation of several Colonies executors. 

## Executors 
| Executor | Purpose                                           |
| :---     | :-----------                                      |
| echo     | An executor that just echo back its input         |
| sleep    | An executor that sleeps, can be used for testings |
| k8s      | Deploys other executors on Kubernetes             |
| backup   | An executor to backup PostgreSQL databases        |

## Colonies functions
### Echo executor
| Function               | Description                             |
| :-----------           | :-----------                            |
| **echo**(text::string) | Closes the process with the text output |

### Sleep executor
| Function                        | Description  |
| :-----------                    | :----------- |
| **sleep**(milliseconds::string) | Sleep        |

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

### Backup executor
| Function     | Description      |
| :----------- | :-----------     |
| **backup**() | Trigger a backup |
