package host

import (
	"net/http"

	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/models"
	"github.com/pkg/errors"
)

type validationID models.HostValidationID

const (
	IsConnected                                    = validationID(models.HostValidationIDConnected)
	HasInventory                                   = validationID(models.HostValidationIDHasDashInventory)
	IsMachineCidrDefined                           = validationID(models.HostValidationIDMachineDashCidrDashDefined)
	BelongsToMachineCidr                           = validationID(models.HostValidationIDBelongsDashToDashMachineDashCidr)
	HasMinCPUCores                                 = validationID(models.HostValidationIDHasDashMinDashCPUDashCores)
	HasMinValidDisks                               = validationID(models.HostValidationIDHasDashMinDashValidDashDisks)
	HasMinMemory                                   = validationID(models.HostValidationIDHasDashMinDashMemory)
	HasCPUCoresForRole                             = validationID(models.HostValidationIDHasDashCPUDashCoresDashForDashRole)
	HasMemoryForRole                               = validationID(models.HostValidationIDHasDashMemoryDashForDashRole)
	IsHostnameUnique                               = validationID(models.HostValidationIDHostnameDashUnique)
	IsHostnameValid                                = validationID(models.HostValidationIDHostnameDashValid)
	IsAPIVipConnected                              = validationID(models.HostValidationIDAPIDashVipDashConnected)
	BelongsToMajorityGroup                         = validationID(models.HostValidationIDBelongsDashToDashMajorityDashGroup)
	IsPlatformNetworkSettingsValid                 = validationID(models.HostValidationIDValidDashPlatformDashNetworkDashSettings)
	IsNTPSynced                                    = validationID(models.HostValidationIDNtpDashSynced)
	SucessfullOrUnknownContainerImagesAvailability = validationID(models.HostValidationIDContainerDashImagesDashAvailable)
	AreLsoRequirementsSatisfied                    = validationID(models.HostValidationIDLsoDashRequirementsDashSatisfied)
	AreOcsRequirementsSatisfied                    = validationID(models.HostValidationIDOcsDashRequirementsDashSatisfied)
	AreCnvRequirementsSatisfied                    = validationID(models.HostValidationIDCnvDashRequirementsDashSatisfied)
	SufficientOrUnknownInstallationDiskSpeed       = validationID(models.HostValidationIDSufficientDashInstallationDashDiskDashSpeed)
	HasSufficientNetworkLatencyRequirementForRole  = validationID(models.HostValidationIDSufficientDashNetworkDashLatencyDashRequirementDashForDashRole)
	HasSufficientPacketLossRequirementForRole      = validationID(models.HostValidationIDSufficientDashPacketDashLossDashRequirementDashForDashRole)
	HasDefaultRoute                                = validationID(models.HostValidationIDHasDashDefaultDashRoute)
	IsAPIDomainNameResolvedCorrectly               = validationID(models.HostValidationIDAPIDashDomainDashNameDashResolvedDashCorrectly)
	IsAPIInternalDomainNameResolvedCorrectly       = validationID(models.HostValidationIDAPIDashIntDashDomainDashNameDashResolvedDashCorrectly)
	IsAppsDomainNameResolvedCorrectly              = validationID(models.HostValidationIDAppsDashDomainDashNameDashResolvedDashCorrectly)
	CompatibleWithClusterPlatform                  = validationID(models.HostValidationIDCompatibleDashWithDashClusterDashPlatform)
	IsDNSWildcardNotConfigured                     = validationID(models.HostValidationIDDNSDashWildcardDashNotDashConfigured)
	DiskEncryptionRequirementsSatisfied            = validationID(models.HostValidationIDDiskDashEncryptionDashRequirementsDashSatisfied)
)

func (v validationID) category() (string, error) {
	switch v {
	case IsConnected,
		IsMachineCidrDefined,
		BelongsToMachineCidr,
		IsAPIVipConnected,
		BelongsToMajorityGroup,
		IsNTPSynced,
		SucessfullOrUnknownContainerImagesAvailability,
		HasSufficientNetworkLatencyRequirementForRole,
		HasSufficientPacketLossRequirementForRole,
		HasDefaultRoute,
		IsAPIDomainNameResolvedCorrectly,
		IsAPIInternalDomainNameResolvedCorrectly,
		IsPlatformNetworkSettingsValid,
		IsAppsDomainNameResolvedCorrectly,
		IsDNSWildcardNotConfigured:
		return "network", nil
	case HasInventory,
		HasMinCPUCores,
		HasMinValidDisks,
		HasMinMemory,
		SufficientOrUnknownInstallationDiskSpeed,
		HasCPUCoresForRole,
		HasMemoryForRole,
		IsHostnameUnique,
		IsHostnameValid,
		CompatibleWithClusterPlatform,
		DiskEncryptionRequirementsSatisfied:
		return "hardware", nil
	case AreLsoRequirementsSatisfied,
		AreOcsRequirementsSatisfied,
		AreCnvRequirementsSatisfied:
		return "operators", nil
	}
	return "", common.NewApiError(http.StatusInternalServerError, errors.Errorf("Unexpected validation id %s", string(v)))
}

func (v validationID) String() string {
	return string(v)
}
