package internals

import (
	"fmt"
	"strings"
)

const (
	MachineImageAlpine = "struckchure/alpine"
	MachineImageUbuntu = "struckchure/ubuntu"
)

var AllowedMachineImages = []string{MachineImageAlpine, MachineImageUbuntu}

func ValidateMachineImage(image string) error {
	image = strings.TrimSpace(image)
	for _, allowed := range AllowedMachineImages {
		if image == allowed {
			return nil
		}
	}
	return fmt.Errorf("machineImage must be %q or %q", MachineImageAlpine, MachineImageUbuntu)
}

func DefaultMachineImage() string {
	return MachineImageAlpine
}
