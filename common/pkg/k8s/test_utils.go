package k8s

func createTestDeploymentSpec() DeploymentSpec {
	spec := DeploymentSpec{
		NumberOfPods:           1,
		ExecutorsPerPod:        2,
		ColoniesTLS:            true,
		ColoniesServerHost:     "colonies.colonyos.io",
		ColoniesServerPort:     443,
		ColoniesColonyID:       "4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4",
		ColoniesColonyPrvKey:   "ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514",
		ColoniesExecutorID:     "3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac",
		ColoniesExecutorPrvKey: "ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05",
		EnableRamdisk:          false,
		RamdiskSize:            "1Gi",
		DockerImage:            "colonyos/sleepexecutor:v1.0.0",
		DockerRegistryURL:      "docker_reg",
		DockerRegistryUsername: "docker_username",
		DockerRegistryPassword: "docker_password",
	}

	return spec
}

func createTestJobSpec() JobSpec {
	spec := JobSpec{
		JobName:           CreateUniqueJobName("test"),
		JobContainerImage: "busybox",
		ExecCmd:           "echo",
		ArgsStr:           "helloworld!",
		Parallelism:       4,
		ContainersPerPod:  2,
		CPU:               "500m",
		Memory:            "1Gi",
		UseGPU:            false,
		GPUCount:          1,
		GPUName:           "nvidia-gtx-2080ti",
		PVCName:           "",
		MountPath:         "/cfs",
		ProcessID:         "test_processid",
	}

	return spec
}

func createTestPVCSpec() *PVCSpec {
	spec := &PVCSpec{
		PVCName:      "kube-executor-pvc",
		StorageClass: "longhorn",
		DiskSize:     "1Gi",
	}

	return spec
}
