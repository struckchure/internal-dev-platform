package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/struckchure/idp/internals"
)

func (m *model) buildMachineForm() {
	group := huh.NewGroup(
		huh.NewInput().
			Key("machineName").
			Title("Machine name").
			Description("Unique name for this machine").
			Placeholder("e.g. dev-box").
			Value(&m.machineName),
		huh.NewInput().
			Key("cpu").
			Title("CPU").
			Description("Kubernetes-style CPU limit").
			Placeholder("500m").
			Value(&m.machineCPU),
		huh.NewInput().
			Key("memory").
			Title("Memory").
			Description("Kubernetes-style memory limit").
			Placeholder("512Mi").
			Value(&m.machineMemory),
		huh.NewSelect[string]().
			Key("machineImage").
			Title("Container image").
			Description("↑/↓ browse · idp base images with git and SSH").
			Options(
				huh.NewOption("Alpine (struckchure/alpine)", internals.MachineImageAlpine),
				huh.NewOption("Ubuntu (struckchure/ubuntu)", internals.MachineImageUbuntu),
			).
			Value(&m.machineImage),
	).Title("Create machine").Description("idp will provision a container with these settings.").WithShowHelp(true)

	m.machineForm = huh.NewForm(group).WithKeyMap(interactiveFormKeyMap())
}

func (m *model) resetMachineFormDefaults() {
	m.machineName = ""
	m.machineCPU = "500m"
	m.machineMemory = "512Mi"
	m.machineImage = internals.DefaultMachineImage()
}
