package k8s

import "encoding/json"

type JobSpec struct {
	TestMode          bool
	JobName           string
	JobContainerName  string
	JobContainerImage string
	Namespace         string
}

func (spec *JobSpec) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (spec *JobSpec) Equals(spec2 *JobSpec) bool {
	if spec2 == nil {
		return false
	}

	if spec.TestMode == spec2.TestMode &&
		spec.JobName == spec2.JobName &&
		spec.JobContainerName == spec2.JobContainerName &&
		spec.JobContainerImage == spec2.JobContainerImage &&
		spec.Namespace == spec2.Namespace {
		return true
	}

	return false
}

func ConvertJSONToJobSpec(jsonString string) (*JobSpec, error) {
	var spec *JobSpec
	err := json.Unmarshal([]byte(jsonString), &spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}
