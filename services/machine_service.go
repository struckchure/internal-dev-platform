package services

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/struckchure/idp/dao"
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
	"github.com/struckchure/idp/types"
)

const (
	CREATE_MACHINE_QUEUE = "create-machine-queue"
	UPDATE_MACHINE_QUEUE = "update-machine-queue"
	DELETE_MACHINE_QUEUE = "delete-machine-queue"
)

type IMachineService interface {
	ListMachines(types.ListMachineArgs) ([]db.MachineModel, error)
	CreateMachine(types.CreateMachineArgs) (*db.MachineModel, error)
	CreateMachineEventHandler(types.CreateMachineEventHandlerArgs) error
	GetMachine(types.GetMachineArgs) (*db.MachineModel, error)
	UpdateMachine(types.UpdateMachineArgs) error
	UpdateMachineEventHandler(types.UpdateMachineArgs) error
	DeleteMachine(types.DeleteMachineArgs) error
	DeleteMachineEventHandler(types.DeleteMachineEventHandlerArgs) error
}

type MachineService struct {
	rmq              internals.RabbitMQ
	containerManager internals.IContainerManager

	machineDAO dao.IMachineDao
}

func (m *MachineService) ListMachines(args types.ListMachineArgs) (machines []db.MachineModel, err error) {
	machines, err = m.machineDAO.ListMachines(types.ListMachineArgs(args))

	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return machines, nil
}

func (m *MachineService) CreateMachine(args types.CreateMachineArgs) (*db.MachineModel, error) {
	machine, err := m.machineDAO.CreateMachine(
		types.CreateMachineArgs{
			OwnerId:       args.OwnerId,
			CPU:           args.CPU,
			Memory:        args.Memory,
			MachineName:   args.MachineName,
			MachineImage:  args.MachineImage,
			MachineStatus: db.MachineStatusCreating,
		})

	if err != nil {
		return nil, internals.TranslateDAOError(err)
	}

	var _payload struct{ Id string } = struct{ Id string }{Id: machine.ID}
	payload, _ := json.Marshal(_payload)

	m.rmq.Publish(internals.PublishArgs{
		Queue:   CREATE_MACHINE_QUEUE,
		Content: string(payload),
	})

	return machine, nil
}

func (m *MachineService) CreateMachineEventHandler(args types.CreateMachineEventHandlerArgs) error {
	machine, err := m.GetMachine(types.GetMachineArgs{Id: &args.Id})
	if err != nil {
		return err
	}

	machineName, _ := machine.MachineName()
	machineImage, _ := machine.MachineImage()
	cpu, _ := machine.CPU()
	memory, _ := machine.Memory()

	container, err := m.containerManager.CreateContainer(
		internals.CreateContainerArgs{
			Replicas: 1,
			Labels: map[string]string{
				"idp-app": strings.ToLower(internals.RandomString(15)),
			},
			Name:   machineName,
			Image:  machineImage,
			CPU:    cpu,
			Memory: memory,
		})

	if err != nil {
		m.machineDAO.DeleteMachine(types.DeleteMachineArgs{Id: machine.ID})

		log.Println(err)

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	_, err = m.machineDAO.UpdateMachine(
		types.UpdateMachineArgs{
			ID:            machine.ID,
			ContainerId:   lo.ToPtr(string(container.GetName())),
			MachineStatus: lo.ToPtr(db.MachineStatusRunning),
		})

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return nil
}

func (m *MachineService) GetMachine(args types.GetMachineArgs) (machine *db.MachineModel, err error) {
	machine, err = m.machineDAO.GetMachine(types.GetMachineArgs(args))

	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return machine, nil
}

func (m *MachineService) UpdateMachine(args types.UpdateMachineArgs) error {
	machine, err := m.machineDAO.GetMachine(types.GetMachineArgs{Id: &args.ID})
	if err != nil {
		log.Println(err)

		return err
	}

	containerId, _ := machine.ContainerID()
	machineImage, _ := machine.MachineImage()

	var _payload struct {
		Id           string
		MachineId    string
		MachineImage string
	} = struct {
		Id           string
		MachineId    string
		MachineImage string
	}{
		Id:           machine.ID,
		MachineId:    containerId,
		MachineImage: machineImage,
	}
	payload, _ := json.Marshal(_payload)

	m.rmq.Publish(internals.PublishArgs{
		Queue:   UPDATE_MACHINE_QUEUE,
		Content: string(payload),
	})

	return nil
}

func (m *MachineService) UpdateMachineEventHandler(args types.UpdateMachineArgs) error {
	m.containerManager.UpdateContainer(internals.UpdateContainerArgs{
		DeploymentName: *args.ContainerId,
		Ports:          *args.Ports,
	})

	_, err := m.machineDAO.UpdateMachine(types.UpdateMachineArgs(args))
	if err != nil {
		log.Println(err)

		return err
	}

	return nil
}

func (m *MachineService) DeleteMachine(args types.DeleteMachineArgs) error {
	machine, err := m.GetMachine(types.GetMachineArgs{Id: &args.Id})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	containerId, _ := machine.ContainerID()

	var _payload struct {
		Id        string
		MachineId string
	} = struct {
		Id        string
		MachineId string
	}{
		Id:        machine.ID,
		MachineId: containerId,
	}
	payload, _ := json.Marshal(_payload)

	m.rmq.Publish(internals.PublishArgs{
		Queue:   DELETE_MACHINE_QUEUE,
		Content: string(payload),
	})

	return nil
}

func (m *MachineService) DeleteMachineEventHandler(args types.DeleteMachineEventHandlerArgs) error {
	err := m.machineDAO.DeleteMachine(types.DeleteMachineArgs{Id: args.Id})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	m.containerManager.DeleteContainer(internals.DeleteContainerArgs{DeploymentName: *args.ContainerId})

	return nil
}

func NewMachineService(
	rmq internals.RabbitMQ,
	containerManager internals.IContainerManager,
	machineDAO dao.IMachineDao,
) IMachineService {
	return &MachineService{
		rmq:              rmq,
		containerManager: containerManager,
		machineDAO:       machineDAO,
	}
}
