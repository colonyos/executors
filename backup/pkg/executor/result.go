package executor

import (
	"encoding/json"
)

type Result struct {
	Filename         string `json:"filename"`
	Bucket           string `json:"bucket"`
	FilePathS3       string `json:"filepath_s3"`
	ExecTimeBackup   int64  `json:"exectime_backup"`
	Size             int64  `json:"size"`
	SizeS3           int64  `json:"size_s3"`
	ExecTimeUploadS3 int64  `json:"exectime_upload_s3"`
}

func (r *Result) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
