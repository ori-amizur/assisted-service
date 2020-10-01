package host

import (
	"bytes"
	"context"
	"html/template"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/openshift/assisted-service/internal/common"

	"github.com/sirupsen/logrus"

	"github.com/openshift/assisted-service/models"
)

type resetInstallationCmd struct {
	baseCmd
	db *gorm.DB
}

func NewResetInstallationCmd(log logrus.FieldLogger, db *gorm.DB) *resetInstallationCmd {
	return &resetInstallationCmd{
		baseCmd: baseCmd{log: log},
		db:      db,
	}
}

func (h *resetInstallationCmd) GetStep(ctx context.Context, host *models.Host) (*models.Step, error) {
	var cmdStr string
	var cluster common.Cluster
	var err error
	err = h.db.Take(&cluster, "id = ?", host.ClusterID.String()).Error
	if err != nil {
		return nil, err
	}
	if host.Bootstrap {
		cmdStr += "systemctl stop bootkube.service; rm -rf /etc/kubernetes/manifests/* /etc/kubernetes/static-pod-resources/* /opt/openshift/*.done; "
	}
	cmdStr += "/usr/bin/podman rm --all -f; "
	cmdStr += "ip -o -4 addr show | egrep '[ \t]{{.API_VIP}}/|[ \t]{{.INGRESS_VIP}}/' | awk '{ip_del_cmd = sprintf(\"ip addr del %s dev %s\",$4, $2); system(ip_del_cmd);}' ; "
	cmdStr += "systemctl restart agent ; "
	t, err := template.New("cmd").Parse(cmdStr)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	params := map[string]string{
		"API_VIP":     strings.ReplaceAll(cluster.APIVip, ".", "[.]"),
		"INGRESS_VIP": strings.ReplaceAll(cluster.IngressVip, ".", "[.]"),
	}
	if err := t.Execute(buf, params); err != nil {
		return nil, err
	}
	step := &models.Step{}
	step.StepType = models.StepTypeResetInstallation
	step.Command = "bash"
	step.Args = []string{"-c", buf.String()}
	return step, nil
}
