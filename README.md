# Introduction
This repo contains implementation of several Colonies executors. 

## Executors 
| Executor | Purpose                                           |
| :---     | :-----------                                      |
| echo     | An executor that just echo back its input         |
| sleep    | An executor that sleeps, can be used for testings |
| k8s      | Manage Kubernetes                                 |
| backup   | An executor to backup PostgreSQL databases        |
| unix     | Run unix commands                                 |
| python   | Injects and runs Python code                      |

## Functions
### Echo executor
| Function     | Description                                     |
| :----------- | :-----------                                    |
| **echo**     | Closes the process with the input set as output |

### Sleep executor
| Function     | Description  |
| :----------- | :----------- |
| **sleep**    | Sleep        |

### K8s executor
| Function            | Description                       |
| :-----------        | :-----------                      |
| **deploy**          | Deploy an executor on K8s.        |
| **undeploy**        | Undeploy an executor on K8s       |
| **get_deployments** | List deployments                  |
| **scale**           | Scale deployments                 |
| **get_scale**       | Get scale factor of a deployments |
| **get_pods**        | List pods                         |
| **pods**            | Get number of pods                |
| **get_containers**  | List container names              |
| **containers**      | Get number of containers          |
| **restart**         | Restart pod                       |

### Backup executor
| Function     | Description      |
| :----------- | :-----------     |
| **backup**   | Trigger a backup |

### Unix executor
| Function         | Description      |
| :-----------     | :-----------     |
| **unix_command** | Run unix command |

### Python executor
| Function        | Description                                                             |
| :-----------    | :-----------                                                            |
| **python_code** | Injects the python code into the Python session and register a function |

### Machine Learning executor
| Function        | Description                                                             |
| :-----------    | :-----------                                                            |
| **execute**     | Executes python script.                                                 |
