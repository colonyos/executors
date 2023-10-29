package k8s

import "encoding/json"

type PVCSpec struct {
	PVCName      string
	StorageClass string
	DiskSize     string
}

func (spec *PVCSpec) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (spec *PVCSpec) Equals(spec2 *PVCSpec) bool {
	if spec2 == nil {
		return false
	}

	if spec.PVCName == spec2.PVCName &&
		spec.StorageClass == spec2.StorageClass &&
		spec.DiskSize == spec2.DiskSize {
		return true
	}

	return false
}

func ConvertJSONToPVCSpec(jsonString string) (*PVCSpec, error) {
	var spec *PVCSpec
	err := json.Unmarshal([]byte(jsonString), &spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}
