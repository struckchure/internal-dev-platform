package services

import (
	"errors"
	"strings"

	"github.com/struckchure/idp/prisma/db"
)

var ErrMachineNoContainer = errors.New("machine has no container deployment; recreate the machine or wait until it is running")

// MachineDeploymentName returns the K8s deployment name stored on the machine.
func MachineDeploymentName(machine *db.MachineModel) (string, error) {
	if machine == nil {
		return "", ErrMachineNoContainer
	}
	containerID, ok := machine.ContainerID()
	if !ok || strings.TrimSpace(string(containerID)) == "" {
		return "", ErrMachineNoContainer
	}
	return string(containerID), nil
}
