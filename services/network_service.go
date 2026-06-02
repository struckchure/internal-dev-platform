package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/struckchure/idp/dao"
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
	"github.com/struckchure/idp/types"
)

type INetworkService interface {
	ListNetworks(types.ListNetworksArgs) ([]db.NetworkModel, error)
	CreateNetwork(types.CreateNetworkArgs) (*db.NetworkModel, error)
	DeleteNetwork(types.DeleteNetworkArgs) error

	generateHostName(string) string
}

type NetworkService struct {
	env internals.Env

	networkManager   internals.INetworkManager
	containerManager internals.IContainerManager

	networkDAO                 dao.INetworkDao
	machineService             IMachineService
	githubService              IGithubService
	repoConnectionDao          dao.IRepoConnectionDao
	githubAccountConnectionDao dao.IGithubAccountConnectionDao
}

// ListNetwork implements NetworkServiceInterface.
func (n *NetworkService) ListNetworks(args types.ListNetworksArgs) ([]db.NetworkModel, error) {
	networks, err := n.networkDAO.ListNetworks(args)
	if err != nil {
		return nil, internals.HttpError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return networks, nil
}

// CreateNetwork implements NetworkServiceInterface.
func (n *NetworkService) CreateNetwork(args types.CreateNetworkArgs) (*db.NetworkModel, error) {
	machine, err := n.machineService.GetMachine(types.GetMachineArgs{Id: &args.MachineId})
	if err != nil {
		return nil, internals.HttpError{
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
		}
	}

	containerId, _ := machine.ContainerID()
	machineName, _ := machine.MachineName()
	machineImage, _ := machine.MachineImage()

	err = n.machineService.UpdateMachineEventHandler(types.UpdateMachineArgs{
		ID:           machine.ID,
		ContainerId:  &containerId,
		MachineImage: &machineImage,
		Ports: &[]internals.NetworkPort{
			{
				Protocol:        args.Protocol,
				DestinationPort: args.DestinationPort,
			},
		},
	})
	if err != nil {
		return nil, internals.HttpError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	container, err := n.containerManager.GetContainer(internals.GetContainerArgs{DeploymentName: containerId})
	if err != nil {
		return nil, internals.HttpError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	k8sNetwork, err := n.networkManager.CreateNetwork(internals.CreateNetworkArgs{
		Name:     machineName,
		Labels:   container.Spec.Template.Labels,
		HostName: n.generateHostName(machineName),
		Port: internals.NetworkPort{
			Protocol:        args.Protocol,
			DestinationPort: args.DestinationPort,
		},
	})
	if err != nil {
		return nil, internals.HttpError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	network, err := n.networkDAO.CreateNetwork(types.CreateNetworkArgs{
		MachineId:       machine.ID,
		HostName:        k8sNetwork.HostName,
		Protocol:        args.Protocol,
		ListeningPort:   args.ListeningPort,
		DestinationPort: args.DestinationPort,
		ServiceId:       k8sNetwork.ServiceID,
		IngressId:       k8sNetwork.IngressID,
	})
	if err != nil {
		return nil, internals.HttpError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	return network, nil
}

// DeleteNetwork implements NetworkServiceInterface.
func (n *NetworkService) DeleteNetwork(args types.DeleteNetworkArgs) error {
	// TODO: update to use `GetNetwork` from `NetworkServiceInterface`
	network, err := n.networkDAO.GetNetwork(types.GetNetworkArgs{Id: &args.Id})
	if err != nil {
		return internals.TranslateDAOError(err)
	}

	err = n.networkDAO.DeleteNetwork(args)
	if err != nil {
		return internals.TranslateDAOError(err)
	}

	serviceId, _ := network.ServiceID()
	ingressId, _ := network.IngressID()

	// TODO: add k8s error translation
	err = n.networkManager.DeleteNetwork(internals.DeleteNetworkArgs{
		ServiceID: serviceId,
		IngressID: ingressId,
	})
	if err != nil {
		return err
	}

	return nil
}

// generateHostName implements NetworkServiceInterface.
func (n *NetworkService) generateHostName(prefix string) string {
	subDomain := internals.Slugify(fmt.Sprintf("%s-%s", prefix, internals.RandomString(5)))
	fullDomain := fmt.Sprintf("%s.%s", subDomain, n.env.INGRESS_ROOT_DOMAIN)

	return strings.ToLower(fullDomain)
}

func NewNetworkService(
	env internals.Env,

	machineService IMachineService,
	networkDAO dao.INetworkDao,

	containerManager internals.IContainerManager,
	networkManager internals.INetworkManager,
	githubService IGithubService,
	repoConnectionDao dao.IRepoConnectionDao,
	githubAccountConnectionDao dao.IGithubAccountConnectionDao,
) INetworkService {
	return &NetworkService{
		env: env,

		networkManager:   networkManager,
		containerManager: containerManager,

		networkDAO:                 networkDAO,
		machineService:             machineService,
		githubService:              githubService,
		repoConnectionDao:          repoConnectionDao,
		githubAccountConnectionDao: githubAccountConnectionDao,
	}
}
