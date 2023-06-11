# Introduction
This repo contains implementation of several Colonies executors. 

| Executor | Purpose                                            | Colonies function                                                                                                        |
| :---     | :-----------                                       | :-----------                                                                                                             |
| backup   | An executor to backup PostgreSQL databases.        | **backup**()                                                                                                             |
| sleep    | An executor that sleeps. Can be used for testings. | **sleep**(milliseconds::string)                                                                                          |
| k8s      | Deploys other executors on Kubernetes.             | **deploy**(name::string, pods::int, executorperpod::int, ramdisk::bool, dockerimage::string)\ **undeploy**(name::string  |
