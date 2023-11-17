package parsers

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func ValidateFuncSpec(funcSpec *core.FunctionSpec) error {
	if funcSpec.Conditions.ProcessesPerNode == 0 {
		return errors.New("Invalid funcspec, processes-per-node cannot be 0")
	}

	if funcSpec.Conditions.Nodes == 0 {
		return errors.New("Invalid funcspec, nodes cannot be 0")
	}

	if funcSpec.Conditions.WallTime == 0 {
		return errors.New("Invalid funcspec, walltime cannot be 0")
	}

	if funcSpec.MaxExecTime <= 0 {
		return errors.New("Invalid funcspec, maxexectime must be set to a larger value than 0")
	}

	if funcSpec.Conditions.WallTime < int64(funcSpec.MaxExecTime) {
		return errors.New("Invalid funcspec, walltime must be > maxexectime")
	}

	if funcSpec.MaxRetries <= 0 {
		return errors.New("Invalid funcspec, maxretries must be set to a larger value than 0")
	}

	err := ValidateMemory(funcSpec.Conditions.Memory)
	if err != nil {
		return err
	}

	err = ValidateCPU(funcSpec.Conditions.CPU)
	if err != nil {
		return err
	}

	return nil
}
