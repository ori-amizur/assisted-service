package bminventory

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/openshift/assisted-service/pkg/auth"

	"github.com/openshift/assisted-service/internal/identity"

	"github.com/danielerez/go-dns-client/pkg/dnsproviders"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/openshift/assisted-service/internal/cluster"
	"github.com/openshift/assisted-service/internal/cluster/validations"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/internal/events"
	"github.com/openshift/assisted-service/internal/host"
	"github.com/openshift/assisted-service/internal/installcfg"
	"github.com/openshift/assisted-service/internal/metrics"
	"github.com/openshift/assisted-service/internal/network"
	"github.com/openshift/assisted-service/models"
	"github.com/openshift/assisted-service/pkg/filemiddleware"
	"github.com/openshift/assisted-service/pkg/generator"
	"github.com/openshift/assisted-service/pkg/job"
	logutil "github.com/openshift/assisted-service/pkg/log"
	"github.com/openshift/assisted-service/pkg/requestid"
	"github.com/openshift/assisted-service/pkg/s3wrapper"
	"github.com/openshift/assisted-service/pkg/transaction"
	"github.com/openshift/assisted-service/restapi"
	"github.com/openshift/assisted-service/restapi/operations/installer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeconfig = "kubeconfig"

const (
	ResourceKindHost    = "Host"
	ResourceKindCluster = "Cluster"
)

const DefaultUser = "kubeadmin"
const ConsoleUrlPrefix = "https://console-openshift-console.apps"

var (
	DefaultClusterNetworkCidr       = "10.128.0.0/14"
	DefaultClusterNetworkHostPrefix = int64(23)
	DefaultServiceNetworkCidr       = "172.30.0.0/16"
)

type Config struct {
	ImageBuilder        string            `envconfig:"IMAGE_BUILDER" default:"quay.io/ocpmetal/installer-image-build:latest"`
	AgentDockerImg      string            `envconfig:"AGENT_DOCKER_IMAGE" default:"quay.io/ocpmetal/assisted-installer-agent:latest"`
	IgnitionGenerator   string            `envconfig:"IGNITION_GENERATE_IMAGE" default:"quay.io/ocpmetal/assisted-ignition-generator:latest"` // TODO: update the latest once the repository has git workflow
	ServiceBaseURL      string            `envconfig:"SERVICE_BASE_URL"`
	S3EndpointURL       string            `envconfig:"S3_ENDPOINT_URL" default:"http://10.35.59.36:30925"`
	S3Bucket            string            `envconfig:"S3_BUCKET" default:"test"`
	ImageExpirationTime time.Duration     `envconfig:"IMAGE_EXPIRATION_TIME" default:"60m"`
	AwsAccessKeyID      string            `envconfig:"AWS_ACCESS_KEY_ID" default:"accessKey1"`
	AwsSecretAccessKey  string            `envconfig:"AWS_SECRET_ACCESS_KEY" default:"verySecretKey1"`
	DeployTarget        string            `envconfig:"DEPLOY_TARGET" default:"k8s"`
	BaseDNSDomains      map[string]string `envconfig:"BASE_DNS_DOMAINS" default:""`
}

const agentMessageOfTheDay = `
**  **  **  **  **  **  **  **  **  **  **  **  **  **  **  **  **  ** **  **  **  **  **  **  **
This is a host being installed by the OpenShift Assisted Installer.
It will be installed from scratch during the installation.
The primary service is agent.service.  To watch its status run e.g
sudo journalctl -u agent.service
**  **  **  **  **  **  **  **  **  **  **  **  **  **  **  **  **  ** **  **  **  **  **  **  **
`

const ignitionConfigFormat = `{
"ignition": { "version": "2.2.0" },
  "passwd": {
    "users": [
      {{.userSshKey}}
    ]
  },
"systemd": {
"units": [{
"name": "agent.service",
"enabled": true,
"contents": "[Service]\nType=simple\nRestart=always\nRestartSec=3\nStartLimitIntervalSec=0\nEnvironment=HTTP_PROXY={{.HTTPProxy}}\nEnvironment=http_proxy={{.HTTPProxy}}\nEnvironment=HTTPS_PROXY={{.HTTPSProxy}}\nEnvironment=https_proxy={{.HTTPSProxy}}\nEnvironment=NO_PROXY={{.NoProxy}}\nEnvironment=no_proxy={{.NoProxy}}\nEnvironment=PULL_SECRET_TOKEN={{.PullSecretToken}}\nExecStartPre=podman run --privileged --rm -v /usr/local/bin:/hostbin {{.AgentDockerImg}} cp /usr/bin/agent /hostbin\nExecStart=/usr/local/bin/agent --url {{.ServiceBaseURL}} --cluster-id {{.clusterId}} --agent-version {{.AgentDockerImg}}\n\n[Install]\nWantedBy=multi-user.target"
}]
},
"storage": {
    "files": [{
      "filesystem": "root",
      "path": "/etc/motd",
      "mode": 644,
      "contents": { "source": "data:,{{.AGENT_MOTD}}" }
    }]
  }
}`

var clusterFileNames = []string{
	"kubeconfig",
	"bootstrap.ign",
	"master.ign",
	"worker.ign",
	"metadata.json",
	"kubeadmin-password",
	"kubeconfig-noingress",
	"install-config.yaml",
}

type debugCmd struct {
	cmd    string
	stepID string
}

type bareMetalInventory struct {
	Config
	db            *gorm.DB
	debugCmdMap   map[strfmt.UUID]debugCmd
	debugCmdMux   sync.Mutex
	log           logrus.FieldLogger
	hostApi       host.API
	clusterApi    cluster.API
	eventsHandler events.Handler
	s3Client      s3wrapper.API
	metricApi     metrics.API
	generator     generator.ISOInstallConfigGenerator
}

var _ restapi.InstallerAPI = &bareMetalInventory{}

func NewBareMetalInventory(
	db *gorm.DB,
	log logrus.FieldLogger,
	hostApi host.API,
	clusterApi cluster.API,
	cfg Config,
	generator generator.ISOInstallConfigGenerator,
	eventsHandler events.Handler,
	s3Client s3wrapper.API,
	metricApi metrics.API,
) *bareMetalInventory {

	b := &bareMetalInventory{
		db:            db,
		log:           log,
		Config:        cfg,
		debugCmdMap:   make(map[strfmt.UUID]debugCmd),
		hostApi:       hostApi,
		clusterApi:    clusterApi,
		generator:     generator,
		eventsHandler: eventsHandler,
		s3Client:      s3Client,
		metricApi:     metricApi,
	}

	if b.Config.DeployTarget == "k8s" {
		//Run first ISO dummy for image pull, this is done so that the image will be pulled and the api will take less time.
		b.generateDummyISOImage()
	}
	return b
}

func (b *bareMetalInventory) generateDummyISOImage() {
	var (
		dummyId   = "00000000-0000-0000-0000-000000000000"
		jobName   = fmt.Sprintf("dummyimage-%s-%s", dummyId, time.Now().Format("20060102150405"))
		imgName   = fmt.Sprintf("discovery-image-%s", dummyId)
		requestID = requestid.NewID()
		log       = requestid.RequestIDLogger(b.log, requestID)
		cluster   common.Cluster
	)
	// create dummy job without uploading to s3, we just need to pull the image
	if err := b.generator.GenerateISO(requestid.ToContext(context.Background(), requestID), cluster, jobName, imgName, job.Dummy, b.eventsHandler); err != nil {
		log.WithError(err).Errorf("failed to generate dummy ISO image")
	}
}

func (b *bareMetalInventory) formatIgnitionFile(cluster *common.Cluster, params installer.GenerateClusterISOParams) (string, error) {
	creds, err := validations.ParsePullSecret(cluster.PullSecret)
	if err != nil {
		return "", err
	}
	r, ok := creds["cloud.openshift.com"]
	if !ok {
		return "", fmt.Errorf("Pull secret does not contain auth for cloud.openshift.com")
	}

	var ignitionParams = map[string]string{
		"userSshKey":      b.getUserSshKey(params),
		"AgentDockerImg":  b.AgentDockerImg,
		"ServiceBaseURL":  strings.TrimSpace(b.ServiceBaseURL),
		"clusterId":       cluster.ID.String(),
		"PullSecretToken": r.AuthRaw,
		"AGENT_MOTD":      url.PathEscape(agentMessageOfTheDay),
		"HTTPProxy":       cluster.HTTPProxy,
		"HTTPSProxy":      cluster.HTTPSProxy,
		"NoProxy":         cluster.NoProxy,
	}
	tmpl, err := template.New("ignitionConfig").Parse(ignitionConfigFormat)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err = tmpl.Execute(buf, ignitionParams); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (b *bareMetalInventory) getUserSshKey(params installer.GenerateClusterISOParams) string {
	sshKey := params.ImageCreateParams.SSHPublicKey
	if sshKey == "" {
		return ""
	}
	return fmt.Sprintf(`{
		"name": "core",
		"passwordHash": "$6$MWO4bibU8TIWG0XV$Hiuj40lWW7pHiwJmXA8MehuBhdxSswLgvGxEh8ByEzeX2D1dk87JILVUYS4JQOP45bxHRegAB9Fs/SWfszXa5.",
		"sshAuthorizedKeys": [
		"%s"],
		"groups": [ "sudo" ]}`, sshKey)
}

func (b *bareMetalInventory) RegisterCluster(ctx context.Context, params installer.RegisterClusterParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	id := strfmt.UUID(uuid.New().String())
	url := installer.GetClusterURL{ClusterID: id}
	log.Infof("Register cluster: %s with id %s", swag.StringValue(params.NewClusterParams.Name), id)

	if params.NewClusterParams.ClusterNetworkCidr == nil {
		params.NewClusterParams.ClusterNetworkCidr = &DefaultClusterNetworkCidr
	}
	if params.NewClusterParams.ClusterNetworkHostPrefix == 0 {
		params.NewClusterParams.ClusterNetworkHostPrefix = DefaultClusterNetworkHostPrefix
	}
	if params.NewClusterParams.ServiceNetworkCidr == nil {
		params.NewClusterParams.ServiceNetworkCidr = &DefaultServiceNetworkCidr
	}

	cluster := common.Cluster{Cluster: models.Cluster{
		ID:                       &id,
		Href:                     swag.String(url.String()),
		Kind:                     swag.String(ResourceKindCluster),
		BaseDNSDomain:            params.NewClusterParams.BaseDNSDomain,
		ClusterNetworkCidr:       swag.StringValue(params.NewClusterParams.ClusterNetworkCidr),
		ClusterNetworkHostPrefix: params.NewClusterParams.ClusterNetworkHostPrefix,
		IngressVip:               params.NewClusterParams.IngressVip,
		Name:                     swag.StringValue(params.NewClusterParams.Name),
		OpenshiftVersion:         swag.StringValue(params.NewClusterParams.OpenshiftVersion),
		ServiceNetworkCidr:       swag.StringValue(params.NewClusterParams.ServiceNetworkCidr),
		SSHPublicKey:             params.NewClusterParams.SSHPublicKey,
		UpdatedAt:                strfmt.DateTime{},
		UserID:                   auth.UserIDFromContext(ctx),
		OrgID:                    auth.OrgIDFromContext(ctx),
		HTTPProxy:                swag.StringValue(params.NewClusterParams.HTTPProxy),
		HTTPSProxy:               swag.StringValue(params.NewClusterParams.HTTPSProxy),
		NoProxy:                  swag.StringValue(params.NewClusterParams.NoProxy),
		VipDhcpAllocation:        swag.Bool(false),
	}}
	if params.NewClusterParams.PullSecret != "" {
		err := validations.ValidatePullSecret(params.NewClusterParams.PullSecret)
		if err != nil {
			log.WithError(err).Errorf("Pull-secret for new cluster has invalid format")
			return installer.NewRegisterClusterBadRequest().
				WithPayload(common.GenerateError(http.StatusBadRequest, errors.New("Pull-secret has invalid format")))
		}
		setPullSecret(&cluster, params.NewClusterParams.PullSecret)
	}
	if err := validations.ValidateClusterNameFormat(swag.StringValue(params.NewClusterParams.Name)); err != nil {
		return common.NewApiError(http.StatusBadRequest, err)
	}

	err := b.clusterApi.RegisterCluster(ctx, &cluster)
	if err != nil {
		log.Errorf("failed to register cluster %s ", swag.StringValue(params.NewClusterParams.Name))
		return installer.NewRegisterClusterInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	b.metricApi.ClusterRegistered(swag.StringValue(params.NewClusterParams.OpenshiftVersion))
	return installer.NewRegisterClusterCreated().WithPayload(&cluster.Cluster)
}

func (b *bareMetalInventory) DeregisterCluster(ctx context.Context, params installer.DeregisterClusterParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster
	log.Infof("Deregister cluster id %s", params.ClusterID)

	if err := b.db.First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		return installer.NewDeregisterClusterNotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	if err := b.deleteDNSRecordSets(ctx, cluster); err != nil {
		log.Warnf("failed to delete DNS record sets for base domain: %s", cluster.BaseDNSDomain)
	}

	err := b.clusterApi.DeregisterCluster(ctx, &cluster)
	if err != nil {
		log.WithError(err).Errorf("failed to deregister cluster cluster %s", params.ClusterID)
		return installer.NewDeregisterClusterNotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	return installer.NewDeregisterClusterNoContent()
}

func (b *bareMetalInventory) DownloadClusterISO(ctx context.Context, params installer.DownloadClusterISOParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster

	if err := b.db.First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to get cluster %s", params.ClusterID)
		return common.NewApiError(http.StatusNotFound, err)
	}

	imgName := getImageName(*cluster.ID)
	exists, err := b.s3Client.DoesObjectExist(ctx, imgName)
	if err != nil {
		log.WithError(err).Errorf("Failed to get ISO for cluster %s", cluster.ID.String())
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError,
			"Failed to download image: error fetching from storage backend", time.Now())
		return installer.NewDownloadClusterISOInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}
	if !exists {
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError,
			"Failed to download image: the image was not found (perhaps it expired) - please generate the image and try again", time.Now())
		return installer.NewDownloadClusterISONotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, errors.New("The image was not found "+
				"(perhaps it expired) - please generate the image and try again")))
	}
	reader, contentLength, err := b.s3Client.Download(ctx, imgName)
	if err != nil {
		log.WithError(err).Errorf("Failed to get ISO for cluster %s", cluster.ID.String())
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError,
			"Failed to download image: error fetching from storage backend", time.Now())
		return installer.NewDownloadClusterISOInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}
	b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityInfo, "Started image download", time.Now())

	return filemiddleware.NewResponder(installer.NewDownloadClusterISOOK().WithPayload(reader),
		fmt.Sprintf("cluster-%s-discovery.iso", params.ClusterID.String()),
		contentLength)
}

func (b *bareMetalInventory) updateImageInfoPostUpload(ctx context.Context, cluster *common.Cluster) error {
	updates := map[string]interface{}{}
	imgName := getImageName(*cluster.ID)
	imgSize, err := b.s3Client.GetObjectSizeBytes(ctx, imgName)
	if err != nil {
		return errors.New("Failed to generate image: error fetching size")
	}
	updates["image_size_bytes"] = imgSize
	cluster.ImageInfo.SizeBytes = &imgSize

	// Presigned URL only works with AWS S3 because Scality is not exposed
	if b.s3Client.IsAwsS3() {
		signedURL, err := b.s3Client.GeneratePresignedDownloadURL(ctx, imgName, b.Config.ImageExpirationTime)
		if err != nil {
			return errors.New("Failed to generate image: error generating URL")
		}
		updates["image_download_url"] = signedURL
		cluster.ImageInfo.DownloadURL = signedURL
	}

	dbReply := b.db.Model(&models.Cluster{}).Where("id = ?", cluster.ID.String()).Updates(updates)
	if dbReply.Error != nil {
		return errors.New("Failed to generate image: error updating image record")
	}

	return nil
}

func (b *bareMetalInventory) GenerateClusterISO(ctx context.Context, params installer.GenerateClusterISOParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	log.Infof("prepare image for cluster %s", params.ClusterID)
	var cluster common.Cluster

	txSuccess := false
	tx := b.db.Begin()
	defer func() {
		if !txSuccess {
			log.Error("generate cluster ISO failed")
			tx.Rollback()
		}
		if r := recover(); r != nil {
			log.Error("generate cluster ISO failed")
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		msg := "Failed to generate image: error starting DB transaction"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		log.WithError(tx.Error).Errorf("failed to start db transaction")
		return installer.NewInstallClusterInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, errors.New("DB error, failed to start transaction")))
	}

	if err := tx.First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to get cluster: %s", params.ClusterID)
		return installer.NewGenerateClusterISONotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	/* We need to ensure that the metadata in the DB matches the image that will be uploaded to S3,
	so we check that at least 10 seconds have past since the previous request to reduce the chance
	of a race between two consecutive requests.
	*/
	now := time.Now()
	previousCreatedAt := time.Time(cluster.ImageInfo.CreatedAt)
	if previousCreatedAt.Add(10 * time.Second).After(now) {
		log.Error("request came too soon after previous request")
		msg := "Failed to generate image: another request to generate an image has been recently submitted - please wait a few seconds and try again"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewGenerateClusterISOConflict().WithPayload(common.GenerateError(http.StatusConflict,
			errors.New("Another request to generate an image has been recently submitted. Please wait a few seconds and try again.")))
	}

	if !cluster.PullSecretSet {
		errMsg := "Can't generate cluster ISO without pull secret"
		log.Error(errMsg)
		return installer.NewGenerateClusterISOBadRequest().
			WithPayload(common.GenerateError(http.StatusBadRequest, errors.New(errMsg)))
	}

	/* If the request has the same parameters as the previous request and the image is still in S3,
	just refresh the timestamp.
	*/
	var imageExists bool
	if cluster.ImageInfo.SSHPublicKey == params.ImageCreateParams.SSHPublicKey &&
		cluster.ImageInfo.GeneratorVersion == b.Config.ImageBuilder {
		var err error
		imgName := getImageName(params.ClusterID)
		imageExists, err = b.s3Client.UpdateObjectTag(ctx, imgName, "create_sec_since_epoch", strconv.FormatInt(now.Unix(), 10))
		if err != nil {
			log.WithError(err).Errorf("failed to contact storage backend")
			msg := "Failed to generate image: error contacting storage backend"
			b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
			return installer.NewInstallClusterInternalServerError().
				WithPayload(common.GenerateError(http.StatusInternalServerError, errors.New("failed to contact storage backend")))
		}
	}

	updates := map[string]interface{}{}
	updates["image_ssh_public_key"] = params.ImageCreateParams.SSHPublicKey
	updates["image_created_at"] = strfmt.DateTime(now)
	updates["image_expires_at"] = strfmt.DateTime(now.Add(b.Config.ImageExpirationTime))
	updates["image_generator_version"] = b.Config.ImageBuilder
	updates["image_download_url"] = ""
	dbReply := tx.Model(&common.Cluster{}).Where("id = ?", cluster.ID.String()).Updates(updates)
	if dbReply.Error != nil {
		log.WithError(dbReply.Error).Errorf("failed to update cluster: %s", params.ClusterID)
		msg := "Failed to generate image: error updating metadata"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewGenerateClusterISOInternalServerError()
	}

	if err := tx.Commit().Error; err != nil {
		log.Error(err)
		msg := "Failed to generate image: error committing the transaction"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewGenerateClusterISOInternalServerError()
	}
	txSuccess = true
	if err := b.db.Preload("Hosts").First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to get cluster %s after update", params.ClusterID)
		msg := "Failed to generate image: error fetching updated cluster metadata"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewUpdateClusterInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	if imageExists {
		if err := b.updateImageInfoPostUpload(ctx, &cluster); err != nil {
			return installer.NewGenerateClusterISOInternalServerError().
				WithPayload(common.GenerateError(http.StatusInternalServerError, err))
		}

		log.Infof("Re-used existing cluster <%s> image", params.ClusterID)
		b.eventsHandler.AddEvent(ctx, cluster.ID.String(), models.EventSeverityInfo, "Re-used existing image rather than generating a new one", time.Now())
		return installer.NewGenerateClusterISOCreated().WithPayload(&cluster.Cluster)
	}
	ignitionConfig, formatErr := b.formatIgnitionFile(&cluster, params)
	if formatErr != nil {
		log.WithError(formatErr).Errorf("failed to format ignition config file for cluster %s", cluster.ID)
		msg := "Failed to generate image: error formatting ignition file"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewGenerateClusterISOInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, formatErr))
	}

	jobName := fmt.Sprintf("createimage-%s-%s", cluster.ID, now.Format("20060102150405"))
	imgName := getImageName(params.ClusterID)

	if err := b.generator.GenerateISO(ctx, cluster, jobName, imgName, ignitionConfig, b.eventsHandler); err != nil {
		log.WithError(err).Errorf("GenerateISO failed for cluster %s", cluster.ID)
		msg := "Failed to generate image: error in generator.GenerateISO"
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewGenerateClusterISOInternalServerError().WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	if err := b.updateImageInfoPostUpload(ctx, &cluster); err != nil {
		return installer.NewGenerateClusterISOInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	log.Infof("Generated cluster <%s> image with ignition config %s", params.ClusterID, ignitionConfig)
	msg := fmt.Sprintf("Generated image (proxy URL is \"%s\", ", cluster.HTTPProxy)
	if params.ImageCreateParams.SSHPublicKey != "" {
		msg += "SSH public key is set)"
	} else {
		msg += "SSH public key is not set)"
	}
	b.eventsHandler.AddEvent(ctx, cluster.ID.String(), models.EventSeverityInfo, msg, time.Now())
	return installer.NewGenerateClusterISOCreated().WithPayload(&cluster.Cluster)
}

func getImageName(clusterID strfmt.UUID) string {
	return fmt.Sprintf("discovery-image-%s", clusterID.String())
}

type clusterInstaller struct {
	ctx    context.Context
	b      *bareMetalInventory
	log    logrus.FieldLogger
	params installer.InstallClusterParams
}

func (c *clusterInstaller) installHosts(cluster *common.Cluster, tx *gorm.DB) error {
	success := true
	err := errors.Errorf("Failed to install cluster <%s>", cluster.ID.String())
	for i := range cluster.Hosts {
		if installErr := c.b.hostApi.Install(c.ctx, cluster.Hosts[i], tx); installErr != nil {
			success = false
			// collect multiple errors
			err = errors.Wrap(installErr, err.Error())
		}
	}
	if !success {
		return common.NewApiError(http.StatusConflict, err)
	}
	return nil
}

func (b *bareMetalInventory) refreshAllHosts(ctx context.Context, cluster *common.Cluster) error {
	for _, chost := range cluster.Hosts {
		if swag.StringValue(chost.Status) != host.HostStatusKnown {
			return common.NewApiError(http.StatusBadRequest, errors.Errorf("Host %s is in status %s and not ready for install", chost.ID.String(),
				swag.StringValue(chost.Status)))
		}
		err := b.hostApi.RefreshStatus(ctx, chost, b.db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c clusterInstaller) install(tx *gorm.DB) error {
	var cluster common.Cluster
	var err error

	// in case host monitor already updated the state we need to use FOR UPDATE option
	transaction.AddForUpdateQueryOption(tx)

	if err = tx.Preload("Hosts").First(&cluster, "id = ?", c.params.ClusterID).Error; err != nil {
		return errors.Wrapf(err, "failed to find cluster %s", c.params.ClusterID)
	}

	if err = c.b.createDNSRecordSets(c.ctx, cluster); err != nil {
		return errors.Wrapf(err, "failed to create DNS record sets for base domain: %s", cluster.BaseDNSDomain)
	}

	if err = c.b.clusterApi.Install(c.ctx, &cluster, tx); err != nil {
		return errors.Wrapf(err, "failed to install cluster %s", cluster.ID.String())
	}

	// set one of the master nodes as bootstrap
	if err = c.b.setBootstrapHost(c.ctx, cluster, tx); err != nil {
		return err
	}

	// move hosts states to installing
	if err = c.installHosts(&cluster, tx); err != nil {
		return err
	}

	return nil
}

func (b *bareMetalInventory) InstallCluster(ctx context.Context, params installer.InstallClusterParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster
	var err error

	if err = b.db.Preload("Hosts", "status <> ?", host.HostStatusDisabled).First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		return common.NewApiError(http.StatusNotFound, err)
	}
	if err = b.refreshAllHosts(ctx, &cluster); err != nil {
		return common.GenerateErrorResponder(err)
	}
	if _, err = b.clusterApi.RefreshStatus(ctx, &cluster, b.db); err != nil {
		return common.GenerateErrorResponder(err)
	}

	// Reload again after refresh
	if err = b.db.Preload("Hosts", "status <> ?", host.HostStatusDisabled).First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		return common.NewApiError(http.StatusNotFound, err)
	}
	// Verify cluster is ready to install
	if ok, reason := b.clusterApi.IsReadyForInstallation(&cluster); !ok {
		return common.NewApiError(http.StatusConflict,
			errors.Errorf("Cluster is not ready for installation, %s", reason))
	}

	// prepare cluster and hosts for installation
	err = b.db.Transaction(func(tx *gorm.DB) error {
		// in case host monitor already updated the state we need to use FOR UPDATE option
		transaction.AddForUpdateQueryOption(tx)

		if err = b.clusterApi.PrepareForInstallation(ctx, &cluster, tx); err != nil {
			return err
		}

		for i := range cluster.Hosts {
			if err = b.hostApi.PrepareForInstallation(ctx, cluster.Hosts[i], tx); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return common.GenerateErrorResponder(err)
	}

	if err = b.db.Preload("Hosts").First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		return common.GenerateErrorResponder(err)
	}

	go func() {
		var err error
		asyncCtx := requestid.ToContext(context.Background(), requestid.FromContext(ctx))

		defer func() {
			if err != nil {
				log.WithError(err).Warn("Cluster install")
				b.clusterApi.HandlePreInstallError(asyncCtx, &cluster, err)
			}
		}()

		if err = b.generateClusterInstallConfig(asyncCtx, cluster); err != nil {
			return
		}

		cInstaller := clusterInstaller{
			ctx:    asyncCtx, // Need a new context for async part
			b:      b,
			log:    log,
			params: params,
		}
		err = b.db.Transaction(cInstaller.install)
		if err == nil {
			//send metric when the installation process has been started
			b.metricApi.InstallationStarted(cluster.OpenshiftVersion)
		}
	}()

	log.Infof("Successfully prepared cluster <%s> for installation", params.ClusterID.String())
	return installer.NewInstallClusterAccepted().WithPayload(&cluster.Cluster)
}

func (b *bareMetalInventory) setBootstrapHost(ctx context.Context, cluster common.Cluster, db *gorm.DB) error {
	log := logutil.FromContext(ctx, b.log)

	// check if cluster already has bootstrap
	for _, h := range cluster.Hosts {
		if h.Bootstrap {
			log.Infof("Bootstrap ID is %s", h.ID)
			return nil
		}
	}

	masterNodesIds, err := b.clusterApi.GetMasterNodesIds(ctx, &cluster, db)
	if err != nil {
		log.WithError(err).Errorf("failed to get cluster %s master node id's", cluster.ID)
		return errors.Wrapf(err, "Failed to get cluster %s master node id's", cluster.ID)
	}
	if len(masterNodesIds) == 0 {
		return errors.Errorf("Cluster have no master hosts that can operate as bootstrap")
	}
	bootstrapId := masterNodesIds[len(masterNodesIds)-1]
	log.Infof("Bootstrap ID is %s", bootstrapId)
	for i := range cluster.Hosts {
		if cluster.Hosts[i].ID.String() == bootstrapId.String() {
			err = b.hostApi.SetBootstrap(ctx, cluster.Hosts[i], true, db)
			if err != nil {
				log.WithError(err).Errorf("failed to update bootstrap host for cluster %s", cluster.ID)
				return errors.Wrapf(err, "Failed to update bootstrap host for cluster %s", cluster.ID)
			}
		}
	}
	return nil
}

func (b *bareMetalInventory) generateClusterInstallConfig(ctx context.Context, cluster common.Cluster) error {
	log := logutil.FromContext(ctx, b.log)

	cfg, err := installcfg.GetInstallConfig(log, &cluster)
	if err != nil {
		log.WithError(err).Errorf("failed to get install config for cluster %s", cluster.ID)
		return errors.Wrapf(err, "failed to get install config for cluster %s", cluster.ID)
	}

	if err := b.generator.GenerateInstallConfig(ctx, cluster, cfg); err != nil {
		log.WithError(err).Errorf("Faled generating kubeconfig files for cluster %s", cluster.ID)
		return err
	}

	return b.clusterApi.SetGeneratorVersion(&cluster, b.Config.IgnitionGenerator, b.db)
}

func (b *bareMetalInventory) refreshClusterHosts(ctx context.Context, cluster *common.Cluster, tx *gorm.DB, log logrus.FieldLogger) error {
	for _, h := range cluster.Hosts {
		var host models.Host
		var err error
		if err = tx.Take(&host, "id = ? and cluster_id = ?",
			h.ID.String(), cluster.ID.String()).Error; err != nil {
			log.WithError(err).Errorf("failed to find host <%s> in cluster <%s>",
				h.ID.String(), cluster.ID.String())
			return common.NewApiError(http.StatusNotFound, err)
		}
		if err = b.hostApi.RefreshStatus(ctx, &host, tx); err != nil {
			log.WithError(err).Errorf("failed to refresh state of host %s cluster %s", *h.ID, cluster.ID.String())
			return common.NewApiError(http.StatusInternalServerError, err)
		}
	}
	return nil
}

func (b *bareMetalInventory) UpdateCluster(ctx context.Context, params installer.UpdateClusterParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster
	var err error
	log.Info("update cluster ", params.ClusterID)

	if swag.StringValue(params.ClusterUpdateParams.PullSecret) != "" {
		err = validations.ValidatePullSecret(*params.ClusterUpdateParams.PullSecret)
		if err != nil {
			log.WithError(err).Errorf("Pull-secret for cluster %s, has invalid format", params.ClusterID)
			return installer.NewUpdateClusterBadRequest().
				WithPayload(common.GenerateError(http.StatusBadRequest, errors.New("Pull-secret has invalid format")))
		}
	}
	if newClusterName := swag.StringValue(params.ClusterUpdateParams.Name); newClusterName != "" {
		if err = validations.ValidateClusterNameFormat(newClusterName); err != nil {
			return common.NewApiError(http.StatusBadRequest, err)
		}
	}

	txSuccess := false
	tx := b.db.Begin()
	defer func() {
		if !txSuccess {
			log.Error("update cluster failed")
			tx.Rollback()
		}
		if r := recover(); r != nil {
			log.Error("update cluster failed")
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		log.WithError(tx.Error).Errorf("failed to start db transaction")
		return installer.NewUpdateClusterInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, errors.New("DB error, failed to start transaction")))
	}

	// in case host monitor already updated the state we need to use FOR UPDATE option
	transaction.AddForUpdateQueryOption(tx)

	if err = tx.Preload("Hosts").First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to get cluster: %s", params.ClusterID)
		return installer.NewUpdateClusterNotFound().WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	if err = b.clusterApi.VerifyClusterUpdatability(&cluster); err != nil {
		log.WithError(err).Errorf("cluster %s can't be updated in current state", params.ClusterID)
		return installer.NewUpdateClusterConflict().WithPayload(common.GenerateError(http.StatusConflict, err))
	}

	if updateClusterConflict := b.validateDNSDomain(params, log); updateClusterConflict != nil {
		return updateClusterConflict
	}

	err = b.updateClusterData(ctx, &cluster, params, tx, log)
	if err != nil {
		return common.GenerateErrorResponder(err)
	}

	err = b.updateHostsData(ctx, params, tx, log)
	if err != nil {
		return common.GenerateErrorResponder(err)
	}

	err = b.updateHostsAndClusterStatus(ctx, &cluster, tx, log)
	if err != nil {
		return common.GenerateErrorResponder(err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Error(err)
		return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, fmt.Errorf("DB error, failed to commit")))
	}
	txSuccess = true

	if proxySettingsChanged(params.ClusterUpdateParams, &cluster) {
		b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityInfo, "Proxy settings changed", time.Now())
	}

	if err := b.db.Preload("Hosts").First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to get cluster %s after update", params.ClusterID)
		return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
	}

	cluster.HostNetworks = calculateHostNetworks(log, &cluster)
	for _, host := range cluster.Hosts {
		if err := b.customizeHost(host); err != nil {
			return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
		}
	}

	return installer.NewUpdateClusterCreated().WithPayload(&cluster.Cluster)
}

func (b *bareMetalInventory) updateNonDhcpNetworkParams(updates map[string]interface{}, cluster *common.Cluster, params installer.UpdateClusterParams, log logrus.FieldLogger) error {
	apiVip := cluster.APIVip
	ingressVip := cluster.IngressVip
	if params.ClusterUpdateParams.APIVip != nil {
		updates["api_vip"] = *params.ClusterUpdateParams.APIVip
		apiVip = *params.ClusterUpdateParams.APIVip
	}
	if params.ClusterUpdateParams.IngressVip != nil {
		updates["ingress_vip"] = *params.ClusterUpdateParams.IngressVip
		ingressVip = *params.ClusterUpdateParams.IngressVip
	}
	if params.ClusterUpdateParams.MachineNetworkCidr != nil {
		err := errors.New("Setting Machine network CIDR is forbidden when cluster is not in vip-dhcp-allocation mode")
		log.WithError(err).Warnf("Set Machine Network CIDR")
		return common.NewApiError(http.StatusBadRequest, err)
	}
	var machineCidr string

	machineCidr, err := network.CalculateMachineNetworkCIDR(apiVip, ingressVip, cluster.Hosts)
	if err != nil {
		log.WithError(err).Errorf("failed to calculate machine network cidr for cluster: %s", params.ClusterID)
		return common.NewApiError(http.StatusBadRequest, err)
	}
	updates["machine_network_cidr"] = machineCidr

	err = network.VerifyVips(cluster.Hosts, machineCidr, apiVip, ingressVip, false, log)
	if err != nil {
		log.WithError(err).Errorf("VIP verification failed for cluster: %s", params.ClusterID)
		return common.NewApiError(http.StatusBadRequest, err)
	}
	return nil
}

func (b *bareMetalInventory) updateDhcpNetworkParams(updates map[string]interface{}, cluster *common.Cluster, params installer.UpdateClusterParams, log logrus.FieldLogger) error {
	if params.ClusterUpdateParams.APIVip != nil {
		err := errors.New("Setting API VIP is forbidden when cluster is in vip-dhcp-allocation mode")
		log.WithError(err).Warnf("Set API VIP")
		return common.NewApiError(http.StatusBadRequest, err)
	}
	if params.ClusterUpdateParams.IngressVip != nil {
		err := errors.New("Setting Ingress VIP is forbidden when cluster is in vip-dhcp-allocation mode")
		log.WithError(err).Warnf("Set Ingress VIP")
		return common.NewApiError(http.StatusBadRequest, err)
	}
	if params.ClusterUpdateParams.MachineNetworkCidr != nil {
		updates["machine_network_cidr"] = swag.StringValue(params.ClusterUpdateParams.MachineNetworkCidr)
		return network.VerifyMachineCIDR(swag.StringValue(params.ClusterUpdateParams.MachineNetworkCidr), cluster.Hosts, log)
	}
	return nil
}

func (b *bareMetalInventory) updateClusterData(ctx context.Context, cluster *common.Cluster, params installer.UpdateClusterParams, db *gorm.DB, log logrus.FieldLogger) error {
	updates := map[string]interface{}{}
	vipDhcpAllocation := swag.BoolValue(cluster.VipDhcpAllocation)
	if params.ClusterUpdateParams.Name != nil {
		updates["name"] = *params.ClusterUpdateParams.Name
	}
	if params.ClusterUpdateParams.BaseDNSDomain != nil {
		updates["base_dns_domain"] = *params.ClusterUpdateParams.BaseDNSDomain
	}
	if params.ClusterUpdateParams.ClusterNetworkCidr != nil {
		updates["cluster_network_cidr"] = *params.ClusterUpdateParams.ClusterNetworkCidr
	}
	if params.ClusterUpdateParams.ClusterNetworkHostPrefix != nil {
		updates["cluster_network_host_prefix"] = *params.ClusterUpdateParams.ClusterNetworkHostPrefix
	}
	if params.ClusterUpdateParams.ServiceNetworkCidr != nil {
		updates["service_network_cidr"] = *params.ClusterUpdateParams.ServiceNetworkCidr
	}
	if params.ClusterUpdateParams.HTTPProxy != nil {
		updates["http_proxy"] = swag.StringValue(params.ClusterUpdateParams.HTTPProxy)
	}
	if params.ClusterUpdateParams.HTTPSProxy != nil {
		updates["https_proxy"] = swag.StringValue(params.ClusterUpdateParams.HTTPSProxy)
	}
	if params.ClusterUpdateParams.NoProxy != nil {
		updates["no_proxy"] = swag.StringValue(params.ClusterUpdateParams.NoProxy)
	}
	if params.ClusterUpdateParams.VipDhcpAllocation != nil && swag.BoolValue(params.ClusterUpdateParams.VipDhcpAllocation) != vipDhcpAllocation {
		updates["vip_dhcp_allocation"] = swag.BoolValue(params.ClusterUpdateParams.VipDhcpAllocation)
		vipDhcpAllocation = swag.BoolValue(params.ClusterUpdateParams.VipDhcpAllocation)
		if vipDhcpAllocation {
			updates["api_vip"] = ""
			updates["ingress_vip"] = ""
		} else {
			updates["machine_network_cidr"] = ""
		}
	}
	var err error
	if vipDhcpAllocation {
		err = b.updateDhcpNetworkParams(updates, cluster, params, log)
	} else {
		err = b.updateNonDhcpNetworkParams(updates, cluster, params, log)
	}
	if err != nil {
		return err
	}
	if params.ClusterUpdateParams.SSHPublicKey != nil {
		updates["ssh_public_key"] = *params.ClusterUpdateParams.SSHPublicKey
	}

	if params.ClusterUpdateParams.PullSecret != nil {
		cluster.PullSecret = *params.ClusterUpdateParams.PullSecret
		updates["pull_secret"] = *params.ClusterUpdateParams.PullSecret
		if cluster.PullSecret != "" {
			updates["pull_secret_set"] = true
		} else {
			updates["pull_secret_set"] = false
		}
	}

	dbReply := db.Model(&common.Cluster{}).Where("id = ?", cluster.ID.String()).Updates(updates)
	if dbReply.Error != nil {
		log.WithError(dbReply.Error).Errorf("failed to update cluster: %s", params.ClusterID)
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	return nil
}

func (b *bareMetalInventory) updateHostsData(ctx context.Context, params installer.UpdateClusterParams, db *gorm.DB, log logrus.FieldLogger) error {
	for i := range params.ClusterUpdateParams.HostsRoles {
		log.Infof("Update host %s to role: %s", params.ClusterUpdateParams.HostsRoles[i].ID,
			params.ClusterUpdateParams.HostsRoles[i].Role)
		var host models.Host
		err := db.First(&host, "id = ? and cluster_id = ?",
			params.ClusterUpdateParams.HostsRoles[i].ID, params.ClusterID).Error
		if err != nil {
			log.WithError(err).Errorf("failed to find host <%s> in cluster <%s>",
				params.ClusterUpdateParams.HostsRoles[i].ID, params.ClusterID)
			return common.NewApiError(http.StatusNotFound, err)
		}
		err = b.hostApi.UpdateRole(ctx, &host, models.HostRole(params.ClusterUpdateParams.HostsRoles[i].Role), db)
		if err != nil {
			log.WithError(err).Errorf("failed to set role <%s> host <%s> in cluster <%s>",
				params.ClusterUpdateParams.HostsRoles[i].Role, params.ClusterUpdateParams.HostsRoles[i].ID,
				params.ClusterID)
			return common.NewApiError(http.StatusInternalServerError, err)
		}
	}

	for i := range params.ClusterUpdateParams.HostsNames {
		log.Infof("Update host %s to request hostname %s", params.ClusterUpdateParams.HostsNames[i].ID,
			params.ClusterUpdateParams.HostsNames[i].Hostname)
		var host models.Host
		err := db.First(&host, "id = ? and cluster_id = ?",
			params.ClusterUpdateParams.HostsNames[i].ID, params.ClusterID).Error
		if err != nil {
			log.WithError(err).Errorf("failed to find host <%s> in cluster <%s>",
				params.ClusterUpdateParams.HostsRoles[i].ID, params.ClusterID)
			return common.NewApiError(http.StatusNotFound, err)
		}
		err = b.hostApi.UpdateHostname(ctx, &host, params.ClusterUpdateParams.HostsNames[i].Hostname, db)
		if err != nil {
			log.WithError(err).Errorf("failed to set hostname <%s> host <%s> in cluster <%s>",
				params.ClusterUpdateParams.HostsNames[i].Hostname, params.ClusterUpdateParams.HostsNames[i].ID,
				params.ClusterID)
			return common.NewApiError(http.StatusConflict, err)
		}
	}

	return nil
}

func (b *bareMetalInventory) updateHostsAndClusterStatus(ctx context.Context, cluster *common.Cluster, db *gorm.DB, log logrus.FieldLogger) error {
	err := b.refreshClusterHosts(ctx, cluster, db, log)
	if err != nil {
		return err
	}

	if _, err = b.clusterApi.RefreshStatus(ctx, cluster, db); err != nil {
		log.WithError(err).Errorf("failed to validate or update cluster %s state", cluster.ID)
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	return nil
}

func calculateHostNetworks(log logrus.FieldLogger, cluster *common.Cluster) []*models.HostNetwork {
	cidrHostsMap := make(map[string][]strfmt.UUID)
	for _, h := range cluster.Hosts {
		if h.Inventory == "" {
			continue
		}
		var inventory models.Inventory
		err := json.Unmarshal([]byte(h.Inventory), &inventory)
		if err != nil {
			log.WithError(err).Warnf("Could not parse inventory of host %s", *h.ID)
			continue
		}
		for _, intf := range inventory.Interfaces {
			for _, ipv4Address := range intf.IPV4Addresses {
				_, ipnet, err := net.ParseCIDR(ipv4Address)
				if err != nil {
					log.WithError(err).Warnf("Could not parse CIDR %s", ipv4Address)
					continue
				}
				cidr := ipnet.String()
				cidrHostsMap[cidr] = append(cidrHostsMap[cidr], *h.ID)
			}
		}
	}
	ret := make([]*models.HostNetwork, 0)
	for k, v := range cidrHostsMap {
		ret = append(ret, &models.HostNetwork{
			Cidr:    k,
			HostIds: v,
		})
	}
	return ret
}

func (b *bareMetalInventory) ListClusters(ctx context.Context, params installer.ListClustersParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var clusters []*common.Cluster
	query := identity.GetUserIDFilter(ctx)
	if err := b.db.Preload("Hosts").Find(&clusters).Where(query).Error; err != nil {
		log.WithError(err).Error("failed to list clusters")
		return installer.NewListClustersInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}
	var mClusters []*models.Cluster = make([]*models.Cluster, len(clusters))
	for i, c := range clusters {
		mClusters[i] = &c.Cluster
	}

	return installer.NewListClustersOK().WithPayload(mClusters)
}

func (b *bareMetalInventory) GetCluster(ctx context.Context, params installer.GetClusterParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster
	if err := b.db.Preload("Hosts").First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		// TODO: check for the right error
		return installer.NewGetClusterNotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	cluster.HostNetworks = calculateHostNetworks(log, &cluster)
	for _, host := range cluster.Hosts {
		if err := b.customizeHost(host); err != nil {
			return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
		}
	}

	return installer.NewGetClusterOK().WithPayload(&cluster.Cluster)
}

func (b *bareMetalInventory) RegisterHost(ctx context.Context, params installer.RegisterHostParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var host models.Host
	var cluster common.Cluster
	log.Infof("Register host: %+v", params)

	if err := b.db.First(&cluster, "id = ?", params.ClusterID.String()).Error; err != nil {
		log.WithError(err).Errorf("failed to get cluster: %s", params.ClusterID.String())
		if gorm.IsRecordNotFoundError(err) {
			return common.NewApiError(http.StatusNotFound, err)
		}
		return common.NewApiError(http.StatusInternalServerError, err)
	}
	err := b.db.First(&host, "id = ? and cluster_id = ?", *params.NewHostParams.HostID, params.ClusterID).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		log.WithError(err).Errorf("failed to get host %s in cluster: %s",
			*params.NewHostParams.HostID, params.ClusterID.String())
		return installer.NewRegisterHostInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	// In case host doesn't exists check if the cluster accept new hosts registration
	if err != nil && gorm.IsRecordNotFoundError(err) {
		if err := b.clusterApi.AcceptRegistration(&cluster); err != nil {
			log.WithError(err).Errorf("failed to register host <%s> to cluster %s due to: %s",
				params.NewHostParams.HostID, params.ClusterID.String(), err.Error())
			b.eventsHandler.AddEvent(ctx, params.NewHostParams.HostID.String(), models.EventSeverityError,
				"Failed to register host: cluster cannot accept new hosts in its current state", time.Now(), params.ClusterID.String())
			return installer.NewRegisterHostForbidden().
				WithPayload(common.GenerateError(http.StatusForbidden, err))
		}
	}

	url := installer.GetHostURL{ClusterID: params.ClusterID, HostID: *params.NewHostParams.HostID}
	host = models.Host{
		ID:                    params.NewHostParams.HostID,
		Href:                  swag.String(url.String()),
		Kind:                  swag.String(ResourceKindHost),
		ClusterID:             params.ClusterID,
		CheckedInAt:           strfmt.DateTime(time.Now()),
		DiscoveryAgentVersion: params.NewHostParams.DiscoveryAgentVersion,
	}

	if err := b.hostApi.RegisterHost(ctx, &host); err != nil {
		log.WithError(err).Errorf("failed to register host <%s> cluster <%s>",
			params.NewHostParams.HostID.String(), params.ClusterID.String())
		b.eventsHandler.AddEvent(ctx, params.NewHostParams.HostID.String(), models.EventSeverityError,
			"Failed to register host: error creating host metadata", time.Now(), params.ClusterID.String())
		return installer.NewRegisterHostBadRequest().
			WithPayload(common.GenerateError(http.StatusBadRequest, err))
	}

	if err := b.customizeHost(&host); err != nil {
		b.eventsHandler.AddEvent(ctx, params.NewHostParams.HostID.String(), models.EventSeverityError,
			"Failed to register host: error setting host properties", time.Now(), params.ClusterID.String())
		return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
	}

	b.eventsHandler.AddEvent(ctx, params.NewHostParams.HostID.String(), models.EventSeverityInfo,
		fmt.Sprintf("Host %s: registered to cluster", common.GetHostnameForMsg(&host)),
		time.Now(), params.ClusterID.String())
	return installer.NewRegisterHostCreated().WithPayload(&host)
}

func (b *bareMetalInventory) DeregisterHost(ctx context.Context, params installer.DeregisterHostParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	log.Infof("Deregister host: %s cluster %s", params.HostID, params.ClusterID)

	if err := b.db.Where("id = ? and cluster_id = ?", params.HostID, params.ClusterID).
		Delete(&models.Host{}).Error; err != nil {
		// TODO: check error type
		return installer.NewDeregisterHostBadRequest().
			WithPayload(common.GenerateError(http.StatusBadRequest, err))
	}

	// TODO: need to check that host can be deleted from the cluster
	b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityInfo,
		fmt.Sprintf("Host %s: deregistered from cluster", params.HostID.String()), time.Now(), params.ClusterID.String())
	return installer.NewDeregisterHostNoContent()
}

func (b *bareMetalInventory) GetHost(ctx context.Context, params installer.GetHostParams) middleware.Responder {
	var host models.Host
	// TODO: validate what is the error
	if err := b.db.Where("id = ? and cluster_id = ?", params.HostID, params.ClusterID).
		First(&host).Error; err != nil {
		return installer.NewGetHostNotFound().WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	if err := b.customizeHost(&host); err != nil {
		return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
	}

	return installer.NewGetHostOK().WithPayload(&host)
}

func (b *bareMetalInventory) ListHosts(ctx context.Context, params installer.ListHostsParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var hosts []*models.Host
	if err := b.db.Find(&hosts, "cluster_id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to get list of hosts for cluster %s", params.ClusterID)
		return installer.NewListHostsInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	for _, host := range hosts {
		if err := b.customizeHost(host); err != nil {
			return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
		}
	}

	return installer.NewListHostsOK().WithPayload(hosts)
}

func createStepID(stepType models.StepType) string {
	return fmt.Sprintf("%s-%s", stepType, uuid.New().String()[:8])
}

func (b *bareMetalInventory) GetNextSteps(ctx context.Context, params installer.GetNextStepsParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var steps models.Steps
	var host models.Host

	txSuccess := false
	tx := b.db.Begin()
	defer func() {
		if !txSuccess {
			log.Error("get next steps failed")
			tx.Rollback()
		}
		if r := recover(); r != nil {
			log.Error("get next steps failed")
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		log.WithError(tx.Error).Errorf("failed to start db transaction")
		return installer.NewUpdateClusterInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, errors.New("DB error, failed to start transaction")))
	}

	//TODO check the error type
	if err := tx.First(&host, "id = ? and cluster_id = ?", params.HostID, params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to find host: %s", params.HostID)
		return installer.NewGetNextStepsNotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	host.CheckedInAt = strfmt.DateTime(time.Now())
	if err := tx.Model(&host).Update("checked_in_at", host.CheckedInAt).Error; err != nil {
		log.WithError(err).Errorf("failed to update host: %s", params.ClusterID)
		return installer.NewGetNextStepsInternalServerError()
	}

	if err := tx.Commit().Error; err != nil {
		log.Error(err)
		return installer.NewGetNextStepsInternalServerError()
	}
	txSuccess = true

	var err error
	steps, err = b.hostApi.GetNextSteps(ctx, &host)
	if err != nil {
		log.WithError(err).Errorf("failed to get steps for host %s cluster %s", params.HostID, params.ClusterID)
	}

	b.debugCmdMux.Lock()
	if cmd, ok := b.debugCmdMap[params.HostID]; ok {
		step := &models.Step{}
		step.StepType = models.StepTypeExecute
		step.StepID = cmd.stepID
		step.Command = "bash"
		step.Args = []string{"-c", cmd.cmd}
		steps.Instructions = append(steps.Instructions, step)
		delete(b.debugCmdMap, params.HostID)
	}
	b.debugCmdMux.Unlock()

	return installer.NewGetNextStepsOK().WithPayload(&steps)
}

func (b *bareMetalInventory) PostStepReply(ctx context.Context, params installer.PostStepReplyParams) middleware.Responder {
	var err error
	log := logutil.FromContext(ctx, b.log)
	msg := fmt.Sprintf("Received step reply <%s> from cluster <%s> host <%s>  exit-code <%d> stdout <%s> stderr <%s>", params.Reply.StepID, params.ClusterID,
		params.HostID, params.Reply.ExitCode, params.Reply.Output, params.Reply.Error)

	var host models.Host
	if err = b.db.First(&host, "id = ? and cluster_id = ?", params.HostID, params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("Failed to find host <%s> cluster <%s> step <%s> exit code %d stdout <%s> stderr <%s>",
			params.HostID, params.ClusterID, params.Reply.StepID, params.Reply.ExitCode, params.Reply.Output, params.Reply.Error)
		return installer.NewPostStepReplyNotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}

	//check the output exit code
	if params.Reply.ExitCode != 0 {
		err = fmt.Errorf(msg)
		log.WithError(err).Errorf("Exit code is <%d> ", params.Reply.ExitCode)
		handlingError := handleReplyError(params, b, ctx, &host)
		if handlingError != nil {
			log.WithError(handlingError).Errorf("Failed handling reply error for host <%s> cluster <%s>", params.HostID, params.ClusterID)
		}
		return installer.NewPostStepReplyBadRequest().
			WithPayload(common.GenerateError(http.StatusBadRequest, err))
	}

	log.Infof(msg)

	var stepReply string
	stepReply, err = filterReplyByType(params)
	if err != nil {
		log.WithError(err).Errorf("Failed decode <%s> reply for host <%s> cluster <%s>",
			params.Reply.StepID, params.HostID, params.ClusterID)
		return installer.NewPostStepReplyBadRequest().
			WithPayload(common.GenerateError(http.StatusBadRequest, err))
	}

	err = handleReplyByType(params, b, ctx, host, stepReply)
	if err != nil {
		log.WithError(err).Errorf("Failed to update step reply for host <%s> cluster <%s> step <%s>",
			params.HostID, params.ClusterID, params.Reply.StepID)
		return installer.NewPostStepReplyInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	return installer.NewPostStepReplyNoContent()
}

func handleReplyError(params installer.PostStepReplyParams, b *bareMetalInventory, ctx context.Context, h *models.Host) error {

	if params.Reply.StepType == models.StepTypeInstall {
		//if it's install step - need to move host to error
		return b.hostApi.HandleInstallationFailure(ctx, h)
	}
	return nil
}

func (b *bareMetalInventory) updateFreeAddressesReport(ctx context.Context, host *models.Host, freeAddressesReport string) error {
	var (
		err           error
		freeAddresses models.FreeNetworksAddresses
	)
	log := logutil.FromContext(ctx, b.log)
	if err = json.Unmarshal([]byte(freeAddressesReport), &freeAddresses); err != nil {
		log.WithError(err).Warnf("Json unmarshal free addresses of host %s", host.ID.String())
		return err
	}
	if len(freeAddresses) == 0 {
		err = fmt.Errorf("Free addresses for host %s is empty", host.ID.String())
		log.WithError(err).Warn("Update free addresses")
		return err
	}
	if err = b.db.Model(&models.Host{}).Where("id = ? and cluster_id = ?", host.ID.String(),
		host.ClusterID.String()).Updates(map[string]interface{}{"free_addresses": freeAddressesReport}).Error; err != nil {
		log.WithError(err).Warnf("Update free addresses of host %s", host.ID.String())
		return err
	}
	// Gorm sets the number of changed rows in AffectedRows and not the number of matched rows.  Therefore, if the report hasn't changed
	// from the previous report, the AffectedRows will be 0 but it will still be correct.  So no error reporting needed for AffectedRows == 0
	return nil
}

func (b *bareMetalInventory) processDhcpAllocationResponse(ctx context.Context, host *models.Host, dhcpAllocationResponseStr string) error {
	var (
		err                   error
		dhcpAllocationReponse models.DhcpAllocationResponse
		cluster               common.Cluster
	)
	log := logutil.FromContext(ctx, b.log)
	if err = b.db.Take(&cluster, "id = ?", host.ClusterID.String()).Error; err != nil {
		log.WithError(err).Warnf("Get cluster %s", host.ClusterID.String())
		return err
	}
	if !swag.BoolValue(cluster.VipDhcpAllocation) {
		log.Warnf("DHCP not enabled in cluster %s", host.ClusterID.String())
		return nil
	}
	if err = json.Unmarshal([]byte(dhcpAllocationResponseStr), &dhcpAllocationReponse); err != nil {
		log.WithError(err).Warnf("Json unmarshal dhcp allocation from host %s", host.ID.String())
		return err
	}
	apiVip := dhcpAllocationReponse.APIVipAddress.String()
	ingressVip := dhcpAllocationReponse.IngressVipAddress.String()
	isApiVipInMachineCIDR, err := network.IpInCidr(apiVip, cluster.MachineNetworkCidr)
	if err != nil {
		log.WithError(err).Warn("Ip in CIDR for API VIP")
		return err
	}

	isIngressVipInMachineCIDR, err := network.IpInCidr(ingressVip, cluster.MachineNetworkCidr)
	if err != nil {
		log.WithError(err).Warn("Ip in CIDR for Ingress VIP")
		return err
	}

	if !(isApiVipInMachineCIDR && isIngressVipInMachineCIDR) {
		err = errors.Errorf("At least of the IPs (%s, %s) is not in machine CIDR %s", apiVip, ingressVip, cluster.MachineNetworkCidr)
		log.WithError(err).Warn("IP in CIDR")
		return err
	}
	return b.clusterApi.SetVips(ctx, &cluster, apiVip, ingressVip, b.db)
}

func handleReplyByType(params installer.PostStepReplyParams, b *bareMetalInventory, ctx context.Context, host models.Host, stepReply string) error {
	var err error
	switch params.Reply.StepType {
	case models.StepTypeInventory:
		err = b.hostApi.UpdateInventory(ctx, &host, stepReply)
	case models.StepTypeConnectivityCheck:
		err = b.hostApi.UpdateConnectivityReport(ctx, &host, stepReply)
	case models.StepTypeFreeNetworkAddresses:
		err = b.updateFreeAddressesReport(ctx, &host, stepReply)
	case models.StepTypeDhcpLeaseAllocate:
		err = b.processDhcpAllocationResponse(ctx, &host, stepReply)
	}
	return err
}

func filterReplyByType(params installer.PostStepReplyParams) (string, error) {
	var stepReply string
	var err error

	// To make sure we store only information defined in swagger we unmarshal and marshal the stepReplyParams.
	switch params.Reply.StepType {
	case models.StepTypeInventory:
		stepReply, err = filterReply(&models.Inventory{}, params.Reply.Output)
	case models.StepTypeConnectivityCheck:
		stepReply, err = filterReply(&models.ConnectivityReport{}, params.Reply.Output)
	case models.StepTypeFreeNetworkAddresses:
		stepReply, err = filterReply(&models.FreeNetworksAddresses{}, params.Reply.Output)
	case models.StepTypeDhcpLeaseAllocate:
		stepReply, err = filterReply(&models.DhcpAllocationResponse{}, params.Reply.Output)
	}
	return stepReply, err
}

// filterReply return only the expected parameters from the input.
func filterReply(expected interface{}, input string) (string, error) {
	if err := json.Unmarshal([]byte(input), expected); err != nil {
		return "", err
	}
	reply, err := json.Marshal(expected)
	if err != nil {
		return "", err
	}
	return string(reply), nil
}

func (b *bareMetalInventory) SetDebugStep(ctx context.Context, params installer.SetDebugStepParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	stepID := createStepID(models.StepTypeExecute)
	b.debugCmdMux.Lock()
	b.debugCmdMap[params.HostID] = debugCmd{
		cmd:    swag.StringValue(params.Step.Command),
		stepID: stepID,
	}
	b.debugCmdMux.Unlock()
	log.Infof("Added new debug command <%s> for cluster <%s> host <%s>: <%s>",
		stepID, params.ClusterID, params.HostID, swag.StringValue(params.Step.Command))
	b.eventsHandler.AddEvent(ctx, params.ClusterID.String(), models.EventSeverityInfo, "Added debug command", time.Now(), params.HostID.String())
	return installer.NewSetDebugStepNoContent()
}

func (b *bareMetalInventory) DisableHost(ctx context.Context, params installer.DisableHostParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var host models.Host
	log.Info("disabling host: ", params.HostID)

	if err := b.db.First(&host, "id = ? and cluster_id = ?", params.HostID, params.ClusterID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.WithError(err).Errorf("host %s not found", params.HostID)
			return common.NewApiError(http.StatusNotFound, err)
		}
		log.WithError(err).Errorf("failed to get host %s", params.HostID)
		msg := "Failed to disable host: error fetching host from DB"
		b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityError, msg, time.Now(), params.ClusterID.String())
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	if err := b.hostApi.DisableHost(ctx, &host); err != nil {
		log.WithError(err).Errorf("failed to disable host <%s> from cluster <%s>", params.HostID, params.ClusterID)
		msg := "Failed to disable host: error disabling host in current status"
		b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityError, msg, time.Now(), params.ClusterID.String())
		return common.GenerateErrorResponderWithDefault(err, http.StatusConflict)
	}

	if err := b.customizeHost(&host); err != nil {
		msg := "Failed to disable host: error setting host properties"
		b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityError, msg, time.Now(), params.ClusterID.String())
		return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
	}

	msg := "Host disabled by user"
	b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityInfo, msg, time.Now(), params.ClusterID.String())
	return installer.NewDisableHostOK().WithPayload(&host)
}

func (b *bareMetalInventory) EnableHost(ctx context.Context, params installer.EnableHostParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var host models.Host
	log.Info("enable host: ", params.HostID)

	if err := b.db.First(&host, "id = ? and cluster_id = ?", params.HostID, params.ClusterID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.WithError(err).Errorf("host %s not found", params.HostID)
			return common.NewApiError(http.StatusNotFound, err)
		}
		log.WithError(err).Errorf("failed to get host %s", params.HostID)
		msg := "Failed to enable host: error fetching host from DB"
		b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityError, msg, time.Now(), params.ClusterID.String())
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	if err := b.hostApi.EnableHost(ctx, &host); err != nil {
		log.WithError(err).Errorf("failed to enable host <%s> from cluster <%s>", params.HostID, params.ClusterID)
		msg := "Failed to enable host: error disabling host in current status"
		b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityError, msg, time.Now(), params.ClusterID.String())
		return common.GenerateErrorResponderWithDefault(err, http.StatusConflict)
	}

	if err := b.customizeHost(&host); err != nil {
		msg := "Failed to enable host: error setting host properties"
		b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityError, msg, time.Now(), params.ClusterID.String())
		return common.GenerateErrorResponder(common.NewApiError(http.StatusInternalServerError, err))
	}

	msg := "Host enabled by user"
	b.eventsHandler.AddEvent(ctx, params.HostID.String(), models.EventSeverityInfo, msg, time.Now(), params.ClusterID.String())
	return installer.NewEnableHostOK().WithPayload(&host)
}

func (b *bareMetalInventory) GetPresignedForClusterFiles(ctx context.Context, params installer.GetPresignedForClusterFilesParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	// Presigned URL only works with AWS S3 because Scality is not exposed
	if !b.s3Client.IsAwsS3() {
		return common.NewApiError(http.StatusBadRequest, errors.New("Failed to generate presigned URL: invalid backend"))
	}
	if err := b.checkFileForDownload(ctx, params.ClusterID.String(), params.FileName); err != nil {
		return common.GenerateErrorResponder(err)
	}
	duration, _ := time.ParseDuration("10m")
	url, err := b.s3Client.GeneratePresignedDownloadURL(ctx, fmt.Sprintf("%s/%s", params.ClusterID, params.FileName), duration)
	if err != nil {
		log.WithError(err).Errorf("failed to generate presigned URL: %s from cluster: %s", params.FileName, params.ClusterID.String())
		return common.NewApiError(http.StatusInternalServerError, err)
	}
	return installer.NewGetPresignedForClusterFilesOK().WithPayload(&models.Presigned{URL: &url})
}

func (b *bareMetalInventory) DownloadClusterFiles(ctx context.Context, params installer.DownloadClusterFilesParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	if err := b.checkFileForDownload(ctx, params.ClusterID.String(), params.FileName); err != nil {
		return common.GenerateErrorResponder(err)
	}

	respBody, contentLength, err := b.s3Client.Download(ctx, fmt.Sprintf("%s/%s", params.ClusterID, params.FileName))
	if err != nil {
		log.WithError(err).Errorf("failed to download file %s from cluster: %s", params.FileName, params.ClusterID.String())
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	return filemiddleware.NewResponder(installer.NewDownloadClusterFilesOK().WithPayload(respBody), params.FileName, contentLength)
}

func (b *bareMetalInventory) DownloadClusterKubeconfig(ctx context.Context, params installer.DownloadClusterKubeconfigParams) middleware.Responder {
	if err := b.checkFileForDownload(ctx, params.ClusterID.String(), kubeconfig); err != nil {
		return common.GenerateErrorResponder(err)
	}

	respBody, contentLength, err := b.s3Client.Download(ctx, fmt.Sprintf("%s/%s", params.ClusterID, kubeconfig))
	if err != nil {
		return common.NewApiError(http.StatusConflict, err)
	}
	return filemiddleware.NewResponder(installer.NewDownloadClusterKubeconfigOK().WithPayload(respBody), kubeconfig, contentLength)
}

func (b *bareMetalInventory) checkFileForDownload(ctx context.Context, clusterID, fileName string) error {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster
	log.Infof("Checking cluster cluster file for download: %s for cluster %s", fileName, clusterID)

	if !funk.Contains(clusterFileNames, fileName) {
		err := fmt.Errorf("invalid cluster file %s", fileName)
		log.WithError(err).Errorf("failed download file: %s from cluster: %s", fileName, clusterID)
		return common.NewApiError(http.StatusBadRequest, err)
	}

	if err := b.db.First(&cluster, "id = ?", clusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to find cluster %s", clusterID)
		if gorm.IsRecordNotFoundError(err) {
			return common.NewApiError(http.StatusNotFound, err)
		} else {
			return common.NewApiError(http.StatusInternalServerError, err)
		}
	}

	var err error
	if fileName == kubeconfig {
		err = b.clusterApi.DownloadKubeconfig(&cluster)
	} else {
		err = b.clusterApi.DownloadFiles(&cluster)
	}
	if err != nil {
		log.WithError(err).Errorf("failed to get file for cluster %s in current state", clusterID)
		return common.NewApiError(http.StatusConflict, err)
	}
	return nil
}

func (b *bareMetalInventory) GetCredentials(ctx context.Context, params installer.GetCredentialsParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var cluster common.Cluster

	if err := b.db.First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to find cluster %s", params.ClusterID)
		if gorm.IsRecordNotFoundError(err) {
			return common.NewApiError(http.StatusNotFound, err)
		} else {
			return common.NewApiError(http.StatusInternalServerError, err)
		}
	}
	if err := b.clusterApi.GetCredentials(&cluster); err != nil {
		log.WithError(err).Errorf("failed to get credentials of cluster %s", params.ClusterID.String())
		return common.NewApiError(http.StatusConflict, err)
	}
	objectName := fmt.Sprintf("%s/%s", params.ClusterID, "kubeadmin-password")
	r, _, err := b.s3Client.Download(ctx, objectName)
	if err != nil {
		log.WithError(err).Errorf("Failed to get clusters %s object", objectName)
		return common.NewApiError(http.StatusInternalServerError, err)
	}
	defer r.Close()
	password, err := ioutil.ReadAll(r)
	if err != nil {
		log.WithError(fmt.Errorf("%s", password)).Errorf("Failed to get clusters %s", objectName)
		return common.NewApiError(http.StatusConflict, errors.New(string(password)))
	}
	return installer.NewGetCredentialsOK().WithPayload(
		&models.Credentials{
			Username:   DefaultUser,
			Password:   string(password),
			ConsoleURL: fmt.Sprintf("%s.%s.%s", ConsoleUrlPrefix, cluster.Name, cluster.BaseDNSDomain),
		})
}

func (b *bareMetalInventory) UpdateHostInstallProgress(ctx context.Context, params installer.UpdateHostInstallProgressParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	var host models.Host
	if err := b.db.First(&host, "id = ? and cluster_id = ?", params.HostID, params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to find host %s", params.HostID)
		return installer.NewUpdateHostInstallProgressNotFound().
			WithPayload(common.GenerateError(http.StatusNotFound, err))
	}
	if err := b.hostApi.UpdateInstallProgress(ctx, &host, params.HostProgress); err != nil {
		log.WithError(err).Errorf("failed to update host %s progress", params.HostID)
		return installer.NewUpdateHostInstallProgressInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	event := fmt.Sprintf("reached installation stage %s", params.HostProgress.CurrentStage)

	if params.HostProgress.ProgressInfo != "" {
		event += fmt.Sprintf(": %s", params.HostProgress.ProgressInfo)
	}

	log.Info(fmt.Sprintf("Host %s in cluster %s: %s", host.ID, host.ClusterID, event))
	msg := fmt.Sprintf("Host %s: %s", common.GetHostnameForMsg(&host), event)

	b.eventsHandler.AddEvent(ctx, host.ID.String(), models.EventSeverityInfo, msg, time.Now(), host.ClusterID.String())
	return installer.NewUpdateHostInstallProgressOK()
}

func (b *bareMetalInventory) UploadClusterIngressCert(ctx context.Context, params installer.UploadClusterIngressCertParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	log.Infof("UploadClusterIngressCert for cluster %s with params %s", params.ClusterID, params.IngressCertParams)
	var cluster common.Cluster

	if err := b.db.First(&cluster, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to find cluster %s", params.ClusterID)
		if gorm.IsRecordNotFoundError(err) {
			return installer.NewUploadClusterIngressCertNotFound().WithPayload(common.GenerateError(http.StatusNotFound, err))
		} else {
			return installer.NewUploadClusterIngressCertInternalServerError().
				WithPayload(common.GenerateError(http.StatusInternalServerError, err))
		}
	}

	if err := b.clusterApi.UploadIngressCert(&cluster); err != nil {
		return installer.NewUploadClusterIngressCertBadRequest().
			WithPayload(common.GenerateError(http.StatusBadRequest, err))
	}

	objectName := fmt.Sprintf("%s/%s", cluster.ID, kubeconfig)
	exists, err := b.s3Client.DoesObjectExist(ctx, objectName)
	if err != nil {
		log.WithError(err).Errorf("Failed to upload ingress ca")
		return installer.NewUploadClusterIngressCertInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	if exists {
		log.Infof("Ingress ca for cluster %s already exists", cluster.ID)
		return installer.NewUploadClusterIngressCertCreated()
	}

	noingress := fmt.Sprintf("%s/%s-noingress", cluster.ID, kubeconfig)
	resp, _, err := b.s3Client.Download(ctx, noingress)
	if err != nil {
		return installer.NewUploadClusterIngressCertInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	kubeconfigData, err := ioutil.ReadAll(resp)
	if err != nil {
		log.WithError(err).Infof("Failed to convert kubeconfig s3 response to io reader")
		return installer.NewUploadClusterIngressCertInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	mergedKubeConfig, err := mergeIngressCaIntoKubeconfig(kubeconfigData, []byte(params.IngressCertParams), log)
	if err != nil {
		return installer.NewUploadClusterIngressCertInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	if err := b.s3Client.Upload(ctx, mergedKubeConfig, objectName); err != nil {
		return installer.NewUploadClusterIngressCertInternalServerError().
			WithPayload(common.GenerateError(http.StatusInternalServerError, fmt.Errorf("failed to upload %s to s3", objectName)))
	}
	return installer.NewUploadClusterIngressCertCreated()
}

// Merging given ingress ca certificate into kubeconfig
// Code was taken from openshift installer
func mergeIngressCaIntoKubeconfig(kubeconfigData []byte, ingressCa []byte, log logrus.FieldLogger) ([]byte, error) {

	kconfig, err := clientcmd.Load(kubeconfigData)
	if err != nil {
		log.WithError(err).Errorf("Failed to convert kubeconfig data")
		return nil, err
	}
	if kconfig == nil || len(kconfig.Clusters) == 0 {
		err = errors.Errorf("kubeconfig is missing expected data")
		log.Error(err)
		return nil, err
	}

	for _, c := range kconfig.Clusters {
		clusterCABytes := c.CertificateAuthorityData
		if len(clusterCABytes) == 0 {
			err = errors.Errorf("kubeconfig CertificateAuthorityData not found")
			log.Errorf("%e, data %s", err, c.CertificateAuthorityData)
			return nil, err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(clusterCABytes) {
			err = errors.Errorf("cluster CA found in kubeconfig not valid PEM format")
			log.Errorf("%e, ca :%s", err, clusterCABytes)
			return nil, err
		}
		if !certPool.AppendCertsFromPEM(ingressCa) {
			err = errors.Errorf("given ingress-ca is not valid PEM format")
			log.Errorf("%e %s", err, ingressCa)
			return nil, err
		}

		newCA := append(ingressCa, clusterCABytes...)
		c.CertificateAuthorityData = newCA
	}

	kconfigAsByteArray, err := clientcmd.Write(*kconfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert kubeconfig")
	}
	return kconfigAsByteArray, nil
}

func setPullSecret(cluster *common.Cluster, pullSecret string) {
	cluster.PullSecret = pullSecret
	if pullSecret != "" {
		cluster.PullSecretSet = true
	} else {
		cluster.PullSecretSet = false
	}
}

func (b *bareMetalInventory) CancelInstallation(ctx context.Context, params installer.CancelInstallationParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	log.Infof("canceling installation for cluster %s", params.ClusterID)

	var c common.Cluster

	txSuccess := false
	tx := b.db.Begin()
	defer func() {
		if !txSuccess {
			log.Error("cancel installation failed")
			tx.Rollback()
		}
		if r := recover(); r != nil {
			log.Error("cancel installation failed")
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		msg := "Failed to cancel installation: error starting DB transaction"
		log.WithError(tx.Error).Errorf(msg)
		b.eventsHandler.AddEvent(ctx, c.ID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewCancelInstallationInternalServerError().WithPayload(
			common.GenerateError(http.StatusInternalServerError, errors.New(msg)))
	}

	if err := tx.Preload("Hosts").First(&c, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("Failed to cancel installation: could not find cluster %s", params.ClusterID)
		if gorm.IsRecordNotFoundError(err) {
			return installer.NewCancelInstallationNotFound().WithPayload(common.GenerateError(http.StatusNotFound, err))
		}
		return installer.NewCancelInstallationInternalServerError().WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	// cancellation is made by setting the cluster and and hosts states to error.
	if err := b.clusterApi.CancelInstallation(ctx, &c, "Installation was canceled by user", tx); err != nil {
		return common.GenerateErrorResponder(err)
	}
	for _, h := range c.Hosts {
		if err := b.hostApi.CancelInstallation(ctx, h, "Installation was canceled by user", tx); err != nil {
			return common.GenerateErrorResponder(err)
		}
		if err := b.customizeHost(h); err != nil {
			return installer.NewCancelInstallationInternalServerError().WithPayload(common.GenerateError(http.StatusInternalServerError, err))
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("Failed to cancel installation: error committing DB transaction (%s)", err)
		msg := "Failed to cancel installation: error committing DB transaction"
		b.eventsHandler.AddEvent(ctx, c.ID.String(), models.EventSeverityError, msg, time.Now())
		return installer.NewCancelInstallationInternalServerError().WithPayload(
			common.GenerateError(http.StatusInternalServerError, errors.New("DB error, failed to commit transaction")))
	}
	txSuccess = true

	return installer.NewCancelInstallationAccepted().WithPayload(&c.Cluster)
}

func (b *bareMetalInventory) ResetCluster(ctx context.Context, params installer.ResetClusterParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	log.Infof("resetting cluster %s", params.ClusterID)

	var c common.Cluster

	txSuccess := false
	tx := b.db.Begin()
	defer func() {
		if !txSuccess {
			log.Error("reset cluster failed")
			tx.Rollback()
		}
		if r := recover(); r != nil {
			log.Error("reset cluster failed")
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		log.WithError(tx.Error).Errorf("failed to start db transaction")
		return installer.NewResetClusterInternalServerError().WithPayload(
			common.GenerateError(http.StatusInternalServerError, errors.New("DB error, failed to start transaction")))
	}

	if err := tx.Preload("Hosts").First(&c, "id = ?", params.ClusterID).Error; err != nil {
		log.WithError(err).Errorf("failed to find cluster %s", params.ClusterID)
		if gorm.IsRecordNotFoundError(err) {
			return installer.NewResetClusterNotFound().WithPayload(common.GenerateError(http.StatusNotFound, err))
		}
		return installer.NewResetClusterInternalServerError().WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	if err := b.clusterApi.ResetCluster(ctx, &c, "cluster was reset by user", tx); err != nil {
		return common.GenerateErrorResponder(err)
	}

	// abort installation files generation job if running.
	if err := b.generator.AbortInstallConfig(ctx, c); err != nil {
		return installer.NewResetClusterInternalServerError().WithPayload(common.GenerateError(http.StatusInternalServerError, err))
	}

	for _, h := range c.Hosts {
		if err := b.hostApi.ResetHost(ctx, h, "cluster was reset by user", tx); err != nil {
			return common.GenerateErrorResponder(err)
		}
		if err := b.customizeHost(h); err != nil {
			return installer.NewResetClusterInternalServerError().WithPayload(common.GenerateError(http.StatusInternalServerError, err))
		}
	}

	if err := b.deleteS3ClusterFiles(ctx, &c); err != nil {
		return common.NewApiError(http.StatusInternalServerError, err)
	}
	if err := b.deleteDNSRecordSets(ctx, c); err != nil {
		log.Warnf("failed to delete DNS record sets for base domain: %s", c.BaseDNSDomain)
	}

	if err := tx.Commit().Error; err != nil {
		log.Error(err)
		return installer.NewResetClusterInternalServerError().WithPayload(
			common.GenerateError(http.StatusInternalServerError, errors.New("DB error, failed to commit transaction")))
	}
	txSuccess = true

	return installer.NewResetClusterAccepted().WithPayload(&c.Cluster)
}

func (b *bareMetalInventory) CompleteInstallation(ctx context.Context, params installer.CompleteInstallationParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)

	log.Infof("complete cluster %s installation", params.ClusterID)

	var c common.Cluster
	if err := b.db.Preload("Hosts").First(&c, "id = ?", params.ClusterID).Error; err != nil {
		return common.GenerateErrorResponder(err)
	}

	if err := b.clusterApi.CompleteInstallation(ctx, &c, *params.CompletionParams.IsSuccess, params.CompletionParams.ErrorInfo); err != nil {
		log.WithError(err).Errorf("Failed to set complete cluster state on %s ", params.ClusterID.String())
		return common.GenerateErrorResponder(err)
	}

	return installer.NewCompleteInstallationAccepted().WithPayload(&c.Cluster)
}

func (b *bareMetalInventory) deleteS3ClusterFiles(ctx context.Context, c *common.Cluster) error {
	for _, name := range clusterFileNames {
		if err := b.s3Client.DeleteObject(ctx, fmt.Sprintf("%s/%s", c.ID, name)); err != nil {
			return err
		}
	}
	return nil
}

func (b *bareMetalInventory) createDNSRecordSets(ctx context.Context, cluster common.Cluster) error {
	return b.changeDNSRecordSets(ctx, cluster, false)
}

func (b *bareMetalInventory) deleteDNSRecordSets(ctx context.Context, cluster common.Cluster) error {
	return b.changeDNSRecordSets(ctx, cluster, true)
}

func (b *bareMetalInventory) changeDNSRecordSets(ctx context.Context, cluster common.Cluster, delete bool) error {
	log := logutil.FromContext(ctx, b.log)

	domain, err := b.getDNSDomain(cluster.Name, cluster.BaseDNSDomain)
	if err != nil {
		return err
	}
	if domain == nil {
		// No supported base DNS domain specified
		return nil
	}

	switch domain.Provider {
	case "route53":
		var dnsProvider dnsproviders.Provider = dnsproviders.Route53{
			RecordSet: dnsproviders.RecordSet{
				RecordSetType: "A",
				TTL:           60,
			},
			HostedZoneID: domain.ID,
			SharedCreds:  true,
		}

		dnsRecordSetFunc := dnsProvider.CreateRecordSet
		if delete {
			dnsRecordSetFunc = dnsProvider.DeleteRecordSet
		}

		// Create/Delete A record for API Virtual IP
		_, err := dnsRecordSetFunc(domain.APIDomainName, cluster.APIVip)
		if err != nil {
			log.WithError(err).Errorf("failed to update DNS record: (%s, %s)",
				domain.APIDomainName, cluster.APIVip)
			return err
		}
		// Create/Delete A record for Ingress Virtual IP
		_, err = dnsRecordSetFunc(domain.IngressDomainName, cluster.IngressVip)
		if err != nil {
			log.WithError(err).Errorf("failed to update DNS record: (%s, %s)",
				domain.IngressDomainName, cluster.IngressVip)
			return err
		}
		log.Infof("Successfully created DNS records for base domain: %s", cluster.BaseDNSDomain)
	}
	return nil
}

type dnsDomain struct {
	Name              string
	ID                string
	Provider          string
	APIDomainName     string
	IngressDomainName string
}

func (b *bareMetalInventory) getDNSDomain(clusterName, baseDNSDomainName string) (*dnsDomain, error) {
	var dnsDomainID string
	var dnsProvider string

	// Parse base domains from config
	if val, ok := b.Config.BaseDNSDomains[baseDNSDomainName]; ok {
		re := regexp.MustCompile("/")
		if !re.MatchString(val) {
			return nil, errors.New(fmt.Sprintf("Invalid DNS domain: %s", val))
		}
		s := re.Split(val, 2)
		dnsDomainID = s[0]
		dnsProvider = s[1]
	} else {
		// No base domains defined in config
		return nil, nil
	}

	if dnsDomainID == "" || dnsProvider == "" {
		// Specified domain is not defined in config
		return nil, nil
	}

	return &dnsDomain{
		Name:              baseDNSDomainName,
		ID:                dnsDomainID,
		Provider:          dnsProvider,
		APIDomainName:     fmt.Sprintf("%s.%s.%s", "api", clusterName, baseDNSDomainName),
		IngressDomainName: fmt.Sprintf("*.%s.%s.%s", "apps", clusterName, baseDNSDomainName),
	}, nil
}

func (b *bareMetalInventory) validateDNSDomain(params installer.UpdateClusterParams, log logrus.FieldLogger) *installer.UpdateClusterConflict {
	clusterName := swag.StringValue(params.ClusterUpdateParams.Name)
	clusterBaseDomain := swag.StringValue(params.ClusterUpdateParams.BaseDNSDomain)
	dnsDomain, err := b.getDNSDomain(clusterName, clusterBaseDomain)
	if err == nil && dnsDomain != nil {
		// Cluster's baseDNSDomain is defined in config (BaseDNSDomains map)
		if err = b.validateBaseDNS(dnsDomain); err != nil {
			log.WithError(err).Errorf("Invalid base DNS domain: %s", clusterBaseDomain)
			return installer.NewUpdateClusterConflict().
				WithPayload(common.GenerateError(http.StatusConflict,
					errors.New("Base DNS domain isn't configured properly")))
		}
		if err = b.validateDNSRecords(dnsDomain); err != nil {
			log.WithError(err).Errorf("DNS records already exist for cluster: %s", params.ClusterID)
			return installer.NewUpdateClusterConflict().
				WithPayload(common.GenerateError(http.StatusConflict,
					errors.New("DNS records already exist for cluster - please change 'Cluster Name'")))
		}
	}
	return nil
}

func (b *bareMetalInventory) validateBaseDNS(domain *dnsDomain) error {
	return validations.ValidateBaseDNS(domain.Name, domain.ID, domain.Provider)
}

func (b *bareMetalInventory) validateDNSRecords(domain *dnsDomain) error {
	vipAddresses := []string{domain.APIDomainName, domain.IngressDomainName}
	return validations.CheckDNSRecordsExistence(vipAddresses, domain.ID, domain.Provider)
}

func ipAsUint(ipStr string, log logrus.FieldLogger) uint64 {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		log.Warnf("Invalid ip %s", ipStr)
		return 0
	}
	var result uint64 = 0
	for _, p := range parts {
		result = result << 8
		converted, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			log.WithError(err).Warnf("Conversion of %s to uint", p)
			return 0
		}
		result += converted
	}
	return result
}

func applyLimit(ret models.FreeAddressesList, limitParam *int64) models.FreeAddressesList {
	if limitParam != nil && *limitParam >= 0 && *limitParam < int64(len(ret)) {
		return ret[:*limitParam]
	}
	return ret
}

func (b *bareMetalInventory) getFreeAddresses(params installer.GetFreeAddressesParams, log logrus.FieldLogger) (models.FreeAddressesList, error) {
	var hosts []*models.Host
	err := b.db.Select("free_addresses").Find(&hosts, "cluster_id = ? and status in (?)", params.ClusterID.String(), []string{host.HostStatusInsufficient, host.HostStatusKnown}).Error
	if err != nil {
		return nil, common.NewApiError(http.StatusInternalServerError, errors.Wrapf(err, "Error retreiving hosts for cluster %s", params.ClusterID.String()))
	}
	if len(hosts) == 0 {
		return nil, common.NewApiError(http.StatusNotFound, errors.Errorf("No hosts where found for cluster %s", params.ClusterID))
	}
	resultingSet := network.MakeFreeAddressesSet(hosts, params.Network, params.Prefix, log)

	ret := models.FreeAddressesList{}
	for a := range resultingSet {
		ret = append(ret, a)
	}

	// Sort addresses
	sort.Slice(ret, func(i, j int) bool {
		return ipAsUint(ret[i].String(), log) < ipAsUint(ret[j].String(), log)
	})

	ret = applyLimit(ret, params.Limit)

	return ret, nil
}

func (b *bareMetalInventory) GetFreeAddresses(ctx context.Context, params installer.GetFreeAddressesParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)

	results, err := b.getFreeAddresses(params, log)
	if err != nil {
		log.WithError(err).Warn("GetFreeAddresses")
		return common.GenerateErrorResponder(err)
	}
	return installer.NewGetFreeAddressesOK().WithPayload(results)
}

func (b *bareMetalInventory) UploadHostLogs(ctx context.Context, params installer.UploadHostLogsParams) middleware.Responder {
	log := logutil.FromContext(ctx, b.log)
	log.Infof("Uploading logs from host %s in cluster %s", params.HostID, params.ClusterID)

	defer func() {
		// Closing file and removing all temporary files created by Multipart
		params.Upfile.Close()
		params.HTTPRequest.Body.Close()
		err := params.HTTPRequest.MultipartForm.RemoveAll()
		if err != nil {
			log.WithError(err).Warnf("Failed to delete temporary files used for upload")
		}
	}()

	var cluster models.Cluster

	if err := b.db.Preload("Hosts", "id = ?", params.HostID).First(&cluster, "id = ?",
		params.ClusterID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return common.NewApiError(http.StatusNotFound, err)
		}
		return common.NewApiError(http.StatusInternalServerError, err)
	}
	if len(cluster.Hosts) < 1 {
		return common.NewApiError(http.StatusNotFound, errors.Errorf("Host %s not found", params.HostID))
	}
	// needed to get filename
	_, fileHeader, err := params.HTTPRequest.FormFile("upfile")
	if err != nil {
		log.WithError(err).Errorf("Failed to get filename")
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	fileName := fmt.Sprintf("%s/logs/%s/%s", params.ClusterID, common.GetHostnameForMsg(cluster.Hosts[0]), fileHeader.Filename)
	log.Debugf("Start upload %s to bucket %s aws len", fileName, b.S3Bucket)
	err = b.s3Client.UploadStream(ctx, params.Upfile, fileName)

	if err != nil {
		log.WithError(err).Errorf("Failed to upload %s to s3", fileName)
		return common.NewApiError(http.StatusInternalServerError, err)
	}

	log.Infof("Done uploading file %s", fileName)
	return installer.NewUploadHostLogsNoContent()
}

func (b *bareMetalInventory) customizeHost(host *models.Host) error {
	b.customizeHostStages(host)
	b.customizeHostname(host)
	return nil
}

func (b *bareMetalInventory) customizeHostStages(host *models.Host) {
	host.ProgressStages = b.hostApi.GetStagesByRole(host.Role, host.Bootstrap)
}

func (b *bareMetalInventory) customizeHostname(host *models.Host) {
	host.RequestedHostname = common.GetHostnameForMsg(host)
}

func proxySettingsChanged(params *models.ClusterUpdateParams, cluster *common.Cluster) bool {
	if (params.HTTPProxy != nil && cluster.HTTPProxy != swag.StringValue(params.HTTPProxy)) ||
		(params.HTTPSProxy != nil && cluster.HTTPSProxy != swag.StringValue(params.HTTPSProxy)) ||
		(params.NoProxy != nil && cluster.NoProxy != swag.StringValue(params.NoProxy)) {
		return true
	}
	return false
}
