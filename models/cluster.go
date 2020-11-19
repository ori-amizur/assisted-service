// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Cluster cluster
//
// swagger:model cluster
type Cluster struct {

	// The virtual IP used to reach the OpenShift cluster's API.
	// Pattern: ^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3})|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,}))$
	APIVip string `json:"api_vip,omitempty"`

	// The domain name used to reach the OpenShift cluster API.
	APIVipDNSName *string `json:"api_vip_dns_name,omitempty"`

	// Base domain of the cluster. All DNS records must be sub-domains of this base and include the cluster name.
	BaseDNSDomain string `json:"base_dns_domain,omitempty"`

	// IP address block from which Pod IPs are allocated. This block must not overlap with existing physical networks. These IP addresses are used for the Pod network, and if you need to access the Pods from an external network, configure load balancers and routers to manage the traffic.
	// Pattern: ^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-9])|(?:[1-2][0-9])|(?:3[0-2])))|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,})/(?:(?:[0-9])|(?:[1-9][0-9])|(?:1[0-1][0-9])|(?:12[0-8])))$
	ClusterNetworkCidr string `json:"cluster_network_cidr,omitempty"`

	// The subnet prefix length to assign to each individual node. For example, if clusterNetworkHostPrefix is set to 23, then each node is assigned a /23 subnet out of the given cidr (clusterNetworkCIDR), which allows for 510 (2^(32 - 23) - 2) pod IPs addresses. If you are required to provide access to nodes from an external network, configure load balancers and routers to manage the traffic.
	// Maximum: 128
	// Minimum: 1
	ClusterNetworkHostPrefix int64 `json:"cluster_network_host_prefix,omitempty"`

	// Json formatted string containing the majority groups for conectivity checks.
	ConnectivityMajorityGroups string `json:"connectivity_majority_groups,omitempty" gorm:"type:text"`

	// controller logs collected at
	// Format: date-time
	ControllerLogsCollectedAt strfmt.DateTime `json:"controller_logs_collected_at,omitempty" gorm:"type:timestamp with time zone"`

	// The time that this cluster was created.
	// Format: date-time
	CreatedAt strfmt.DateTime `json:"created_at,omitempty" gorm:"type:timestamp with time zone"`

	// The time that the cluster was deleted.
	// Format: date-time
	DeletedAt *strfmt.DateTime `json:"deleted_at,omitempty" gorm:"type:timestamp with time zone"`

	// email domain
	EmailDomain string `json:"email_domain,omitempty"`

	// List of host networks to be filled during query.
	HostNetworks []*HostNetwork `json:"host_networks" gorm:"-"`

	// Hosts that are associated with this cluster.
	Hosts []*Host `json:"hosts" gorm:"foreignkey:ClusterID;association_foreignkey:ID"`

	// Self link.
	// Required: true
	Href *string `json:"href"`

	// A proxy URL to use for creating HTTP connections outside the cluster.
	// http://\<username\>:\<pswd\>@\<ip\>:\<port\>
	//
	HTTPProxy string `json:"http_proxy,omitempty"`

	// A proxy URL to use for creating HTTPS connections outside the cluster.
	// http://\<username\>:\<pswd\>@\<ip\>:\<port\>
	//
	HTTPSProxy string `json:"https_proxy,omitempty" gorm:"column:https_proxy"`

	// Unique identifier of the object.
	// Required: true
	// Format: uuid
	ID *strfmt.UUID `json:"id" gorm:"primary_key"`

	// Json formatted string containing the user overrides for the initial ignition config
	IgnitionConfigOverrides string `json:"ignition_config_overrides,omitempty" gorm:"type:text"`

	// ignition generator version
	IgnitionGeneratorVersion string `json:"ignition_generator_version,omitempty"`

	// image info
	// Required: true
	ImageInfo *ImageInfo `json:"image_info" gorm:"embedded;embedded_prefix:image_"`

	// The virtual IP used for cluster ingress traffic.
	// Pattern: ^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3})|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,}))$
	IngressVip string `json:"ingress_vip,omitempty"`

	// The time that this cluster completed installation.
	// Format: date-time
	InstallCompletedAt strfmt.DateTime `json:"install_completed_at,omitempty" gorm:"type:timestamp with time zone;default:'2000-01-01 00:00:00z'"`

	// JSON-formatted string containing the user overrides for the install-config.yaml file.
	InstallConfigOverrides string `json:"install_config_overrides,omitempty" gorm:"type:text"`

	// The time that this cluster started installation.
	// Format: date-time
	InstallStartedAt strfmt.DateTime `json:"install_started_at,omitempty" gorm:"type:timestamp with time zone;default:'2000-01-01 00:00:00z'"`

	// Indicates the type of this object. Will be 'Cluster' if this is a complete object or 'ClusterLink' if it is just a link,
	// 'AddHostCluster' for cluster that add hosts to existing OCP cluster,
	// 'AddHostsOCPCluster' for cluster running on the OCP and add hosts to it
	//
	// Required: true
	// Enum: [Cluster AddHostsCluster AddHostsOCPCluster]
	Kind *string `json:"kind"`

	// A CIDR that all hosts belonging to the cluster should have an interfaces with IP address that belongs to this CIDR. The api_vip belongs to this CIDR.
	// Pattern: ^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-9])|(?:[1-2][0-9])|(?:3[0-2])))|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,})/(?:(?:[0-9])|(?:[1-9][0-9])|(?:1[0-1][0-9])|(?:12[0-8])))$
	MachineNetworkCidr string `json:"machine_network_cidr,omitempty"`

	// Name of the OpenShift cluster.
	Name string `json:"name,omitempty"`

	// A comma-separated list of destination domain names, domains, IP addresses, or other network CIDRs to exclude from proxying.
	NoProxy string `json:"no_proxy,omitempty"`

	// Cluster ID on OCP system.
	// Format: uuid
	OpenshiftClusterID strfmt.UUID `json:"openshift_cluster_id,omitempty"`

	// Version of the OpenShift cluster.
	// Enum: [4.5 4.6]
	OpenshiftVersion string `json:"openshift_version,omitempty"`

	// org id
	OrgID string `json:"org_id,omitempty"`

	// True if the pull secret has been added to the cluster.
	PullSecretSet bool `json:"pull_secret_set,omitempty"`

	// The IP address pool to use for service IP addresses. You can enter only one IP address pool. If you need to access the services from an external network, configure load balancers and routers to manage the traffic.
	// Pattern: ^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-9])|(?:[1-2][0-9])|(?:3[0-2])))|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,})/(?:(?:[0-9])|(?:[1-9][0-9])|(?:1[0-1][0-9])|(?:12[0-8])))$
	ServiceNetworkCidr string `json:"service_network_cidr,omitempty"`

	// SSH public key for debugging OpenShift nodes.
	SSHPublicKey string `json:"ssh_public_key,omitempty" gorm:"type:varchar(1024)"`

	// Status of the OpenShift cluster.
	// Required: true
	// Enum: [insufficient ready error preparing-for-installation pending-for-input installing finalizing installed adding-hosts cancelled installing-pending-user-action]
	Status *string `json:"status"`

	// Additional information pertaining to the status of the OpenShift cluster.
	// Required: true
	StatusInfo *string `json:"status_info" gorm:"type:varchar(2048)"`

	// The last time that the cluster status was updated.
	// Format: date-time
	StatusUpdatedAt strfmt.DateTime `json:"status_updated_at,omitempty" gorm:"type:timestamp with time zone"`

	// The last time that this cluster was updated.
	// Format: date-time
	UpdatedAt strfmt.DateTime `json:"updated_at,omitempty" gorm:"type:timestamp with time zone"`

	// user name
	UserName string `json:"user_name,omitempty"`

	// JSON-formatted string containing the validation results for each validation id grouped by category (network, hosts-data, etc.)
	ValidationsInfo string `json:"validations_info,omitempty" gorm:"type:varchar(2048)"`

	// Indicate if virtual IP DHCP allocation mode is enabled.
	VipDhcpAllocation *bool `json:"vip_dhcp_allocation,omitempty"`
}

// Validate validates this cluster
func (m *Cluster) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAPIVip(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateClusterNetworkCidr(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateClusterNetworkHostPrefix(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateControllerLogsCollectedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCreatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDeletedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateHostNetworks(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateHosts(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateHref(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateImageInfo(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIngressVip(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInstallCompletedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInstallStartedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateKind(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMachineNetworkCidr(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOpenshiftClusterID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOpenshiftVersion(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateServiceNetworkCidr(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatusInfo(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatusUpdatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUpdatedAt(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Cluster) validateAPIVip(formats strfmt.Registry) error {

	if swag.IsZero(m.APIVip) { // not required
		return nil
	}

	if err := validate.Pattern("api_vip", "body", string(m.APIVip), `^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3})|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,}))$`); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateClusterNetworkCidr(formats strfmt.Registry) error {

	if swag.IsZero(m.ClusterNetworkCidr) { // not required
		return nil
	}

	if err := validate.Pattern("cluster_network_cidr", "body", string(m.ClusterNetworkCidr), `^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-9])|(?:[1-2][0-9])|(?:3[0-2])))|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,})/(?:(?:[0-9])|(?:[1-9][0-9])|(?:1[0-1][0-9])|(?:12[0-8])))$`); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateClusterNetworkHostPrefix(formats strfmt.Registry) error {

	if swag.IsZero(m.ClusterNetworkHostPrefix) { // not required
		return nil
	}

	if err := validate.MinimumInt("cluster_network_host_prefix", "body", int64(m.ClusterNetworkHostPrefix), 1, false); err != nil {
		return err
	}

	if err := validate.MaximumInt("cluster_network_host_prefix", "body", int64(m.ClusterNetworkHostPrefix), 128, false); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateControllerLogsCollectedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.ControllerLogsCollectedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("controller_logs_collected_at", "body", "date-time", m.ControllerLogsCollectedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateCreatedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.CreatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("created_at", "body", "date-time", m.CreatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateDeletedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.DeletedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("deleted_at", "body", "date-time", m.DeletedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateHostNetworks(formats strfmt.Registry) error {

	if swag.IsZero(m.HostNetworks) { // not required
		return nil
	}

	for i := 0; i < len(m.HostNetworks); i++ {
		if swag.IsZero(m.HostNetworks[i]) { // not required
			continue
		}

		if m.HostNetworks[i] != nil {
			if err := m.HostNetworks[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("host_networks" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Cluster) validateHosts(formats strfmt.Registry) error {

	if swag.IsZero(m.Hosts) { // not required
		return nil
	}

	for i := 0; i < len(m.Hosts); i++ {
		if swag.IsZero(m.Hosts[i]) { // not required
			continue
		}

		if m.Hosts[i] != nil {
			if err := m.Hosts[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("hosts" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Cluster) validateHref(formats strfmt.Registry) error {

	if err := validate.Required("href", "body", m.Href); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	if err := validate.FormatOf("id", "body", "uuid", m.ID.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateImageInfo(formats strfmt.Registry) error {

	if err := validate.Required("image_info", "body", m.ImageInfo); err != nil {
		return err
	}

	if m.ImageInfo != nil {
		if err := m.ImageInfo.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("image_info")
			}
			return err
		}
	}

	return nil
}

func (m *Cluster) validateIngressVip(formats strfmt.Registry) error {

	if swag.IsZero(m.IngressVip) { // not required
		return nil
	}

	if err := validate.Pattern("ingress_vip", "body", string(m.IngressVip), `^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3})|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,}))$`); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateInstallCompletedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.InstallCompletedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("install_completed_at", "body", "date-time", m.InstallCompletedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateInstallStartedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.InstallStartedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("install_started_at", "body", "date-time", m.InstallStartedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

var clusterTypeKindPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["Cluster","AddHostsCluster","AddHostsOCPCluster"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		clusterTypeKindPropEnum = append(clusterTypeKindPropEnum, v)
	}
}

const (

	// ClusterKindCluster captures enum value "Cluster"
	ClusterKindCluster string = "Cluster"

	// ClusterKindAddHostsCluster captures enum value "AddHostsCluster"
	ClusterKindAddHostsCluster string = "AddHostsCluster"

	// ClusterKindAddHostsOCPCluster captures enum value "AddHostsOCPCluster"
	ClusterKindAddHostsOCPCluster string = "AddHostsOCPCluster"
)

// prop value enum
func (m *Cluster) validateKindEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, clusterTypeKindPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Cluster) validateKind(formats strfmt.Registry) error {

	if err := validate.Required("kind", "body", m.Kind); err != nil {
		return err
	}

	// value enum
	if err := m.validateKindEnum("kind", "body", *m.Kind); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateMachineNetworkCidr(formats strfmt.Registry) error {

	if swag.IsZero(m.MachineNetworkCidr) { // not required
		return nil
	}

	if err := validate.Pattern("machine_network_cidr", "body", string(m.MachineNetworkCidr), `^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-9])|(?:[1-2][0-9])|(?:3[0-2])))|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,})/(?:(?:[0-9])|(?:[1-9][0-9])|(?:1[0-1][0-9])|(?:12[0-8])))$`); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateOpenshiftClusterID(formats strfmt.Registry) error {

	if swag.IsZero(m.OpenshiftClusterID) { // not required
		return nil
	}

	if err := validate.FormatOf("openshift_cluster_id", "body", "uuid", m.OpenshiftClusterID.String(), formats); err != nil {
		return err
	}

	return nil
}

var clusterTypeOpenshiftVersionPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["4.5","4.6"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		clusterTypeOpenshiftVersionPropEnum = append(clusterTypeOpenshiftVersionPropEnum, v)
	}
}

const (

	// ClusterOpenshiftVersionNr45 captures enum value "4.5"
	ClusterOpenshiftVersionNr45 string = "4.5"

	// ClusterOpenshiftVersionNr46 captures enum value "4.6"
	ClusterOpenshiftVersionNr46 string = "4.6"
)

// prop value enum
func (m *Cluster) validateOpenshiftVersionEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, clusterTypeOpenshiftVersionPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Cluster) validateOpenshiftVersion(formats strfmt.Registry) error {

	if swag.IsZero(m.OpenshiftVersion) { // not required
		return nil
	}

	// value enum
	if err := m.validateOpenshiftVersionEnum("openshift_version", "body", m.OpenshiftVersion); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateServiceNetworkCidr(formats strfmt.Registry) error {

	if swag.IsZero(m.ServiceNetworkCidr) { // not required
		return nil
	}

	if err := validate.Pattern("service_network_cidr", "body", string(m.ServiceNetworkCidr), `^(?:(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-9])|(?:[1-2][0-9])|(?:3[0-2])))|(?:(?:[0-9a-fA-F]*:[0-9a-fA-F]*){2,})/(?:(?:[0-9])|(?:[1-9][0-9])|(?:1[0-1][0-9])|(?:12[0-8])))$`); err != nil {
		return err
	}

	return nil
}

var clusterTypeStatusPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["insufficient","ready","error","preparing-for-installation","pending-for-input","installing","finalizing","installed","adding-hosts","cancelled","installing-pending-user-action"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		clusterTypeStatusPropEnum = append(clusterTypeStatusPropEnum, v)
	}
}

const (

	// ClusterStatusInsufficient captures enum value "insufficient"
	ClusterStatusInsufficient string = "insufficient"

	// ClusterStatusReady captures enum value "ready"
	ClusterStatusReady string = "ready"

	// ClusterStatusError captures enum value "error"
	ClusterStatusError string = "error"

	// ClusterStatusPreparingForInstallation captures enum value "preparing-for-installation"
	ClusterStatusPreparingForInstallation string = "preparing-for-installation"

	// ClusterStatusPendingForInput captures enum value "pending-for-input"
	ClusterStatusPendingForInput string = "pending-for-input"

	// ClusterStatusInstalling captures enum value "installing"
	ClusterStatusInstalling string = "installing"

	// ClusterStatusFinalizing captures enum value "finalizing"
	ClusterStatusFinalizing string = "finalizing"

	// ClusterStatusInstalled captures enum value "installed"
	ClusterStatusInstalled string = "installed"

	// ClusterStatusAddingHosts captures enum value "adding-hosts"
	ClusterStatusAddingHosts string = "adding-hosts"

	// ClusterStatusCancelled captures enum value "cancelled"
	ClusterStatusCancelled string = "cancelled"

	// ClusterStatusInstallingPendingUserAction captures enum value "installing-pending-user-action"
	ClusterStatusInstallingPendingUserAction string = "installing-pending-user-action"
)

// prop value enum
func (m *Cluster) validateStatusEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, clusterTypeStatusPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Cluster) validateStatus(formats strfmt.Registry) error {

	if err := validate.Required("status", "body", m.Status); err != nil {
		return err
	}

	// value enum
	if err := m.validateStatusEnum("status", "body", *m.Status); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateStatusInfo(formats strfmt.Registry) error {

	if err := validate.Required("status_info", "body", m.StatusInfo); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateStatusUpdatedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.StatusUpdatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("status_updated_at", "body", "date-time", m.StatusUpdatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Cluster) validateUpdatedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.UpdatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("updated_at", "body", "date-time", m.UpdatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Cluster) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Cluster) UnmarshalBinary(b []byte) error {
	var res Cluster
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
