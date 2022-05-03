package hostcommands

import (
	"context"
	"sync"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/openshift/assisted-service/internal/connectivity"
	"github.com/openshift/assisted-service/models"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

const maxPerMinute = 150

type connectivityCheckCmd struct {
	baseCmd
	db                     *gorm.DB
	connectivityValidator  connectivity.Validator
	connectivityCheckImage string
	queue                  hashmap.HashMap
}

func NewConnectivityCheckCmd(log logrus.FieldLogger, db *gorm.DB, connectivityValidator connectivity.Validator, connectivityCheckImage string) *connectivityCheckCmd {
	return &connectivityCheckCmd{
		baseCmd:                baseCmd{log: log},
		db:                     db,
		connectivityValidator:  connectivityValidator,
		connectivityCheckImage: connectivityCheckImage,
	}
}

type clusterQueue struct {
	mutex         sync.Mutex
	admittedQueue []time.Time
	waitingQueue  []string
}

func (c *connectivityCheckCmd) isAllowed(host *models.Host) bool {
	valIntf, _ := c.queue.GetOrInsert(host.ClusterID.String(), &clusterQueue{})
	q := valIntf.(*clusterQueue)
	q.mutex.Lock()
	defer q.mutex.Unlock()
	firstValidIndex := 0
	for ; firstValidIndex < len(q.admittedQueue) && time.Since(q.admittedQueue[firstValidIndex]) > time.Minute; firstValidIndex++ {
	}
	q.admittedQueue = q.admittedQueue[firstValidIndex:]
	id := host.ID.String()
	if !funk.ContainsString(q.waitingQueue, id) {
		q.waitingQueue = append(q.waitingQueue, id)
	}
	index := funk.IndexOfString(q.waitingQueue, id)
	if index < (maxPerMinute - len(q.admittedQueue)) {
		q.waitingQueue = append(q.waitingQueue[:index], q.waitingQueue[index+1:]...)
		q.admittedQueue = append(q.admittedQueue, time.Now())
		return true
	}
	return false
}

func (c *connectivityCheckCmd) GetSteps(ctx context.Context, host *models.Host) ([]*models.Step, error) {

	var hosts []*models.Host
	if err := c.db.Select("id", "inventory", "status").Find(&hosts, "cluster_id = ?", host.ClusterID).Error; err != nil {
		c.log.WithError(err).Errorf("failed to get list of hosts for cluster %s", host.ClusterID)
		return nil, err
	}
	if len(hosts) <= maxPerMinute || c.isAllowed(host) {

		hostsData, err := convertHostsToConnectivityCheckParams(host.ID, hosts, c.connectivityValidator)
		if err != nil {
			c.log.WithError(err).Errorf("failed to convert hosts to connectivity params for host %s cluster %s", host.ID, host.ClusterID)
			return nil, err
		}

		// Skip this step in case there is no hosts to check
		if hostsData == "" {
			return nil, nil
		}

		step := &models.Step{
			StepType: models.StepTypeConnectivityCheck,
			Command:  "",
			Args: []string{
				hostsData,
			},
		}
		return []*models.Step{step}, nil
	}
	return nil, nil
}
