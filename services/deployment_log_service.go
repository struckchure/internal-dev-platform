package services

import (
	"github.com/struckchure/idp/dao"
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
	"github.com/struckchure/idp/types"
)

type DeploymentLogService struct {
	deploymentLogDAO dao.IDeploymentLogDao
}

// ListDeploymentLogs implements DeploymentLogServiceInterface.
func (s *DeploymentLogService) ListDeploymentLogs(args types.GetDeploymentArgs) ([]db.DeploymentLogModel, error) {
	deploymentLogs, err := s.deploymentLogDAO.ListLogs(types.ListDeploymentLogArgs{DeploymentId: &args.Id})
	if err != nil {
		return nil, internals.TranslateDAOError(err)
	}

	return deploymentLogs, nil
}

// CreateDeploymentLog implements DeploymentLogServiceInterface.
func (s *DeploymentLogService) CreateDeploymentLog(args types.CreateDeploymentLogArgs) (*db.DeploymentLogModel, error) {
	logExists, err := s.deploymentLogDAO.ListLogs(types.ListDeploymentLogArgs{
		DeploymentId: &args.DeploymentId,
		JobId:        &args.JobId,
		Message:      &args.Message,
	})
	if err != nil {
		return nil, internals.TranslateDAOError(err)
	}

	if len(logExists) > 0 {
		return &logExists[0], nil
	}

	return s.deploymentLogDAO.CreateLog(args)
}

func NewDeploymentLogService(deploymentLogDAO dao.IDeploymentLogDao) *DeploymentLogService {
	return &DeploymentLogService{deploymentLogDAO: deploymentLogDAO}
}
