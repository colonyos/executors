package parsers

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// mib is the format "1000Mi"
// 1 MB is 1000000 bytes and 1 MiB is 1048576 bytes.
func convertKitoM(kib string) (string, error) {
	kibFloat, err := strconv.ParseFloat(strings.TrimSuffix(kib, "Ki"), 64)
	if err != nil {
		return "", err
	}

	mibFloat := kibFloat / 1024

	mbValue := int64(math.Round((mibFloat * 1048576) / 1000000))

	return fmt.Sprintf("%dM", mbValue), nil
}

// mib is the format "1000Mi"
// 1 MB is 1000000 bytes and 1 MiB is 1048576 bytes.
func convertMitoM(mib string) (string, error) {
	mibFloat, err := strconv.ParseFloat(strings.TrimSuffix(mib, "Mi"), 64)
	if err != nil {
		return "", err
	}

	mbValue := int64(math.Round((mibFloat * 1048576) / 1000000))

	return fmt.Sprintf("%dM", mbValue), nil
}

// mib is the format "1000Mi"
// 1 MB is 1000000 bytes and 1 MiB is 1048576 bytes.
func convertGitoM(gib string) (string, error) {
	gibFloat, err := strconv.ParseFloat(strings.TrimSuffix(gib, "Gi"), 64)
	if err != nil {
		return "", err
	}

	mibFloat := gibFloat * 1024

	mbValue := int64(math.Round((mibFloat * 1048576) / 1000000))

	return fmt.Sprintf("%dM", mbValue), nil
}

// Memory is defined in Mi, Gi in function specs, we need to covert that to M
func ParseMemory(memStr string) (string, error) {
	err := ValidateMemory(memStr)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(memStr, "Mi") {
		return convertMitoM(memStr)
	} else if strings.HasSuffix(memStr, "Gi") {
		return convertGitoM(memStr)
	} else if strings.HasSuffix(memStr, "Ki") {
		return convertKitoM(memStr)
	} else {
		return "", errors.New("Failed to convert memory from Mi to M: " + memStr)
	}
}

func ValidateMemory(memStr string) error {
	if len(memStr) < 3 {
		return errors.New("Memory must be specified in the following format: {value}Mi/Gi, e.g.1Mi or 1Gi")
	}

	if strings.HasPrefix(memStr, "-") {
		return errors.New("Memory cannot be negative")
	}

	if strings.HasSuffix(memStr, "Mi") || strings.HasSuffix(memStr, "Gi") || (strings.HasSuffix(memStr, "Mi")) {
		return nil
	}

	return errors.New("Memory must be defined in Mi or Gi")

}
