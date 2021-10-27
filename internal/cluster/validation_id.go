package cluster

import (
	"net/http"

	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/models"
	"github.com/pkg/errors"
)

type ValidationID models.ClusterValidationID

const (
	isClusterCidrDefined                = ValidationID(models.ClusterValidationIDClusterDashCidrDashDefined)
	isServiceCidrDefined                = ValidationID(models.ClusterValidationIDServiceDashCidrDashDefined)
	noCidrOverlapping                   = ValidationID(models.ClusterValidationIDNoDashCidrsDashOverlapping)
	networkPrefixValid                  = ValidationID(models.ClusterValidationIDNetworkDashPrefixDashValid)
	IsMachineCidrDefined                = ValidationID(models.ClusterValidationIDMachineDashCidrDashDefined)
	IsMachineCidrEqualsToCalculatedCidr = ValidationID(models.ClusterValidationIDMachineDashCidrDashEqualsDashToDashCalculatedDashCidr)
	IsApiVipDefined                     = ValidationID(models.ClusterValidationIDAPIDashVipDashDefined)
	IsApiVipValid                       = ValidationID(models.ClusterValidationIDAPIDashVipDashValid)
	isNetworkTypeValid                  = ValidationID(models.ClusterValidationIDNetworkDashTypeDashValid)
	IsIngressVipDefined                 = ValidationID(models.ClusterValidationIDIngressDashVipDashDefined)
	IsIngressVipValid                   = ValidationID(models.ClusterValidationIDIngressDashVipDashValid)
	AllHostsAreReadyToInstall           = ValidationID(models.ClusterValidationIDAllDashHostsDashAreDashReadyDashToDashInstall)
	SufficientMastersCount              = ValidationID(models.ClusterValidationIDSufficientDashMastersDashCount)
	IsDNSDomainDefined                  = ValidationID(models.ClusterValidationIDDNSDashDomainDashDefined)
	IsPullSecretSet                     = ValidationID(models.ClusterValidationIDPullDashSecretDashSet)
	IsNtpServerConfigured               = ValidationID(models.ClusterValidationIDNtpDashServerDashConfigured)
	IsOcsRequirementsSatisfied          = ValidationID(models.ClusterValidationIDOcsDashRequirementsDashSatisfied)
	IsLsoRequirementsSatisfied          = ValidationID(models.ClusterValidationIDLsoDashRequirementsDashSatisfied)
	IsCnvRequirementsSatisfied          = ValidationID(models.ClusterValidationIDCnvDashRequirementsDashSatisfied)
)

func (v ValidationID) Category() (string, error) {
	switch v {
	case IsMachineCidrDefined, IsMachineCidrEqualsToCalculatedCidr, IsApiVipDefined, IsApiVipValid, IsIngressVipDefined,
		IsIngressVipValid, isClusterCidrDefined, isServiceCidrDefined, noCidrOverlapping, networkPrefixValid,
		IsDNSDomainDefined, IsNtpServerConfigured, isNetworkTypeValid:
		return "network", nil
	case AllHostsAreReadyToInstall, SufficientMastersCount:
		return "hosts-data", nil
	case IsPullSecretSet:
		return "configuration", nil
	case IsOcsRequirementsSatisfied, IsLsoRequirementsSatisfied, IsCnvRequirementsSatisfied:
		return "operators", nil
	}
	return "", common.NewApiError(http.StatusInternalServerError, errors.Errorf("Unexpected cluster validation id %s", string(v)))
}

func (v ValidationID) String() string {
	return string(v)
}
