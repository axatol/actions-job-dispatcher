package job

import (
	"fmt"
	"strings"
)

const (
	RunnerLabelSelfHosted = "self-hosted"
	// RunnerLabelPrefixCompute = "cpu-"
	// RunnerLabelPrefixMemory  = "memory-"
)

type RunnerLabels []string

func (rl RunnerLabels) Occurences(key string) int {
	total := 0
	for _, label := range rl {
		if strings.HasPrefix(label, key) {
			total += 1
		}
	}

	return total
}

func (rl RunnerLabels) Validate() error {
	if len(rl) < 1 {
		return fmt.Errorf("must have at least one runner label")
	}

	selfHosted := rl.Occurences(RunnerLabelSelfHosted)
	if selfHosted != 1 {
		return fmt.Errorf("must contain self-hosted runner label")
	}

	// compute := rl.Occurences(RunnerLabelPrefixCompute)
	// if compute > 1 {
	// 	return fmt.Errorf("cannot contain more than one cpu spec label")
	// }

	// memory := rl.Occurences(RunnerLabelPrefixMemory)
	// if memory > 1 {
	// 	return fmt.Errorf("cannot contain more than one memory spec label")
	// }

	// total number of labels leaves no extra label for id
	if len(rl) == selfHosted /* +compute+memory */ {
		return fmt.Errorf("must provide identifying runner label")
	}

	return nil
}

func (rl RunnerLabels) String() string {
	return strings.Join(rl, ",")
}

type RunnerLabelSpec struct {
	Name string
	// Compute string
	// Memory  string
}

func RunnerLabelSpecFromLabels(labels []string) (*RunnerLabelSpec, error) {
	if err := RunnerLabels(labels).Validate(); err != nil {
		return nil, err
	}

	spec := RunnerLabelSpec{}
	for _, label := range labels {
		if label == RunnerLabelSelfHosted {
			continue
		}

		// if value, ok := strings.CutPrefix(label, RunnerLabelPrefixCompute); ok {
		// 	spec.Compute = value
		// 	continue
		// }

		// if value, ok := strings.CutPrefix(label, RunnerLabelPrefixMemory); ok {
		// 	spec.Memory = value
		// 	continue
		// }

		spec.Name = label
	}

	return &spec, nil
}

func (r RunnerLabelSpec) String() string {
	return ""
}
