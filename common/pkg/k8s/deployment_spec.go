package k8s

import "encoding/json"

type DeploymentSpec struct {
	TestMode               bool
	DeploymentName         string
	Namespace              string
	NumberOfPods           int
	ExecutorsPerPod        int
	ColoniesTLS            bool
	ColoniesServerHost     string
	ColoniesServerPort     int
	ColoniesColonyID       string
	ColoniesColonyPrvKey   string
	ColoniesExecutorID     string
	ColoniesExecutorPrvKey string
	EnableRamdisk          bool
	RamdiskSize            string
	DockerImage            string
	DockerRegistryURL      string
	DockerRegistryUsername string
	DockerRegistryPassword string
}

func (spec *DeploymentSpec) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (spec *DeploymentSpec) Equals(spec2 *DeploymentSpec) bool {
	if spec2 == nil {
		return false
	}

	if spec.TestMode == spec2.TestMode &&
		spec.DeploymentName == spec2.DeploymentName &&
		spec.Namespace == spec2.Namespace &&
		spec.NumberOfPods == spec2.NumberOfPods &&
		spec.ExecutorsPerPod == spec2.ExecutorsPerPod &&
		spec.ColoniesServerHost == spec2.ColoniesServerHost &&
		spec.ColoniesServerPort == spec2.ColoniesServerPort &&
		spec.ColoniesColonyID == spec2.ColoniesColonyID &&
		spec.ColoniesColonyPrvKey == spec2.ColoniesColonyPrvKey &&
		spec.RamdiskSize == spec2.RamdiskSize &&
		spec.DockerRegistryURL == spec2.DockerRegistryURL &&
		spec.DockerRegistryUsername == spec2.DockerRegistryUsername &&
		spec.DockerRegistryPassword == spec2.DockerRegistryPassword {
		return true
	}

	return false
}

func ConvertJSONToDeploymentSpec(jsonString string) (*DeploymentSpec, error) {
	var spec *DeploymentSpec
	err := json.Unmarshal([]byte(jsonString), &spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}
