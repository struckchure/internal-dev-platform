package services

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/go-github/v56/github"
	"github.com/samber/lo"
	"github.com/struckchure/idp/dao"
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
	"github.com/struckchure/idp/types"
)

type DeploymentService struct {
	deploymentDAO              dao.IDeploymentDao
	deploymentLogDAO           dao.IDeploymentLogDao
	repoConnectionDao          dao.IRepoConnectionDao
	githubAccountConnectionDao dao.IGithubAccountConnectionDao
	machineDAO                 dao.IMachineDao
	githubService              IGithubService
}

// ListDeployments implements DeploymentServiceInterface.
func (s *DeploymentService) ListDeployments(args types.ListDeploymentArgs) ([]db.DeploymentModel, error) {
	deployments, err := s.deploymentDAO.ListDeployments(args)
	if err != nil {
		return nil, internals.TranslateDAOError(err)
	}

	return deployments, nil
}

func (s *DeploymentService) GetDeployment(args types.GetDeploymentArgs) (*db.DeploymentModel, error) {
	deployment, err := s.deploymentDAO.GetDeployment(args)
	if err != nil {
		return nil, internals.TranslateDAOError(err)
	}

	return deployment, nil
}

type ListBranches struct {
	UserId       string `json:"userId" swaggerignore:"true"`
	ConnectionId string `json:"connectionId"`
}

func (s *DeploymentService) ListBranches(args ListBranches) ([]*github.Branch, error) {
	// TODO: check object owner permission
	repoConnection, err := s.repoConnectionDao.GetRepoConnection(types.GetRepoConnectionArgs{Id: args.ConnectionId})
	if err != nil {
		return nil, err
	}

	accountConnection, err := s.githubAccountConnectionDao.GetConnection(types.GetGithubAccountConnectionsArgs{
		UserId: lo.ToPtr(repoConnection.Machine().OwnerID),
	})
	if err != nil {
		return nil, err
	}

	repoFullName, _ := repoConnection.RepoName()
	installationId, _ := accountConnection.GithubInstallationID()

	accessToken, err := s.githubService.GetInstallationToken(installationId)
	if err != nil {
		return nil, err
	}

	branches, err := s.githubService.ListBranches(types.ListBranchesArgs{
		Token:        *accessToken,
		RepoFullName: repoFullName,
	})

	return branches, err
}

type ListCommitsArgs struct {
	UserId       string `json:"userId" swaggerignore:"true"`
	ConnectionId string `json:"connectionId"`
	Ref          string `json:"ref"`
}

func (s *DeploymentService) ListCommits(args ListCommitsArgs) ([]*github.RepositoryCommit, error) {
	// TODO: check object owner permission
	repoConnection, err := s.repoConnectionDao.GetRepoConnection(types.GetRepoConnectionArgs{Id: args.ConnectionId})
	if err != nil {
		return nil, err
	}

	accountConnection, err := s.githubAccountConnectionDao.GetConnection(types.GetGithubAccountConnectionsArgs{
		UserId: lo.ToPtr(repoConnection.Machine().OwnerID),
	})
	if err != nil {
		return nil, err
	}

	repoFullName, _ := repoConnection.RepoName()
	installationId, _ := accountConnection.GithubInstallationID()

	accessToken, err := s.githubService.GetInstallationToken(installationId)
	if err != nil {
		return nil, err
	}

	return s.githubService.ListCommits(types.ListCommitsArgs{
		Token:        *accessToken,
		Ref:          args.Ref,
		RepoFullName: repoFullName,
	})
}

type DeployRepoArgs struct {
	UserId       string `json:"userId" swaggerignore:"true"`
	ConnectionId string `json:"connectionId"`
	Ref          string `json:"ref"`
}

func (s *DeploymentService) DeployRepo(args DeployRepoArgs) error {
	// TODO: check object owner permission
	repoConnection, err := s.repoConnectionDao.GetRepoConnection(types.GetRepoConnectionArgs{Id: args.ConnectionId})
	if err != nil {
		return internals.TranslateDAOError(err)
	}

	accountConnection, err := s.githubAccountConnectionDao.GetConnection(types.GetGithubAccountConnectionsArgs{
		UserId: lo.ToPtr(args.UserId),
	})
	if err != nil {
		return err
	}

	machine, err := s.machineDAO.GetMachine(types.GetMachineArgs{Id: &repoConnection.MachineID})
	if err != nil {
		return internals.TranslateDAOError(err)
	}
	if _, err := MachineDeploymentName(machine); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	repoFullName, _ := repoConnection.RepoName()
	_repoId, _ := repoConnection.RepoID()
	repoId, _ := strconv.Atoi(_repoId)
	installationId, _ := accountConnection.GithubInstallationID()

	return s.githubService.DeployRepo(types.DeployRepoArgs{
		InstallationId: installationId,
		MachineId:      repoConnection.MachineID,
		RepoId:         repoId,
		RepoFullName:   repoFullName,
		Ref:            args.Ref,
	})
}

func NewDeploymentService(
	deploymentDAO dao.IDeploymentDao,
	deploymentLogDAO dao.IDeploymentLogDao,
	repoConnectionDao dao.IRepoConnectionDao,
	githubAccountConnectionDao dao.IGithubAccountConnectionDao,
	machineDAO dao.IMachineDao,
	githubService IGithubService,
) *DeploymentService {
	return &DeploymentService{
		deploymentDAO:              deploymentDAO,
		deploymentLogDAO:           deploymentLogDAO,
		repoConnectionDao:          repoConnectionDao,
		githubAccountConnectionDao: githubAccountConnectionDao,
		machineDAO:                 machineDAO,
		githubService:              githubService,
	}
}
