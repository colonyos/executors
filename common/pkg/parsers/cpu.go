package parsers

import (
	"errors"
	"strconv"
	"strings"
)

func ParseCPU(millicpu string) (string, error) {
	err := ValidateCPU(millicpu)
	if err != nil {
		return "", err
	}

	millicpu = strings.TrimSuffix(millicpu, "m")

	value, err := strconv.Atoi(millicpu)
	if err != nil {
		return "", err
	}

	if value > 0 && value < 1000 {
		return "1", nil
	}
	return strconv.Itoa((value + 999) / 1000), nil
}

func ValidateCPU(millicpu string) error {
	if len(millicpu) <= 2 {
		return errors.New("CPU must be specified in the following format: {value}m, e.g. 1000m")
	}

	if strings.HasPrefix(millicpu, "-") {
		return errors.New("CPU cannot be negative")
	}

	if !strings.HasSuffix(millicpu, "m") {
		return errors.New("CPU must be specified in the following format: {value}m, e.g. 1000m")
	}

	return nil
}
