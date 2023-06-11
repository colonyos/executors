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
| Description                 | Function                                                                                               |
| :-----------                | :-----------                                                                                           |
| Trigger a backup            | **backup**()                                                                                           |

### Sleep executor
| Description                 | Function                                                                                               |
| :-----------                | :-----------                                                                                           |
| Sleep                       | **sleep**(milliseconds::string)                                                                        |

### K8s executor
| Description                 | Function                                                                                               |
| :-----------                | :-----------                                                                                           |
| Deploy an executor on K8s   | **deploy**(deploymentname::string, pods::int, executorperpod::int, ramdisk::bool, dockerimage::string) |
| Undeploy an executor on K8s | **undeploy**(name::string)                                                                             |
| List deployments            | **get_deployments**()                                                                                  |
| Scale deployments           | **scale**(name::string, pods::int)                                                                     |
| List pods                   | **get_pods**(deploymentname::string)                                                                   |
| Get number of pods          | **pods**(name::string)                                                                                 |
| List container names        | **get_containers**(podname::string)                                                                    |
| Get number of containers    | **containers**(podname::string)                                                                        |
| Restart pod                 | **restart**(podname::string)                                                                           |
| Get output                  | **get_output**(podname::string, containername::string)                                                 |
