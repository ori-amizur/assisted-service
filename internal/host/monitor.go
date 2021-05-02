package host

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/go-openapi/swag"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/models"
	"github.com/openshift/assisted-service/pkg/requestid"
	"github.com/thoas/go-funk"
)

type stats struct {
	elapsed, user, system time.Duration
	calls                 int
}

var durations map[string]stats

func toTime(t syscall.Timeval) time.Time {
	return time.Unix(t.Sec, t.Usec*1000)
}

func timeIt(f func(), label string) {
	var rstart, rend syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &rstart)
	start := time.Now()
	f()
	end := time.Now()
	duration := end.Sub(start)
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &rend)
	current, _ := durations[label]
	current.elapsed = time.Duration(current.elapsed.Nanoseconds() + duration.Nanoseconds())
	current.calls++
	current.user = time.Duration(toTime(rend.Utime).Sub(toTime(rstart.Utime)).Nanoseconds() + current.user.Nanoseconds())
	current.system = time.Duration(toTime(rend.Stime).Sub(toTime(rstart.Stime)).Nanoseconds() + current.system.Nanoseconds())
	durations[label] = current
}

func printTimes() {
	for k, v := range durations {
		fmt.Printf("%s  elapsed %+v ms user %+v ms sys %+v ms calls %d\n", k, v.elapsed.Milliseconds(), v.user.Milliseconds(), v.system.Milliseconds(), v.calls)
	}
}

func (m *Manager) SkipMonitoring(h *models.Host) bool {
	skipMonitoringStates := []string{string(models.LogsStateCompleted), string(models.LogsStateTimeout), ""}
	result := ((swag.StringValue(h.Status) == models.HostStatusError || swag.StringValue(h.Status) == models.HostStatusCancelled) &&
		funk.Contains(skipMonitoringStates, h.LogsInfo))
	return result
}

func (m *Manager) HostMonitoring() {
	if !m.leaderElector.IsLeader() {
		m.log.Debugf("Not a leader, exiting HostMonitoring")
		return
	}
	m.log.Debugf("Running HostMonitoring")
	var (
		offset    int
		limit     = m.Config.MonitorBatchSize
		requestID = requestid.NewID()
		ctx       = requestid.ToContext(context.Background(), requestID)
		log       = requestid.RequestIDLogger(m.log, requestID)
		err       error
	)

	monitorStates := []string{
		models.HostStatusDiscovering,
		models.HostStatusKnown,
		models.HostStatusDisconnected,
		models.HostStatusInsufficient,
		models.HostStatusPendingForInput,
		models.HostStatusPreparingForInstallation,
		models.HostStatusPreparingSuccessful,
		models.HostStatusInstalling,
		models.HostStatusInstallingInProgress,
		models.HostStatusInstalled,
		models.HostStatusInstallingPendingUserAction,
		models.HostStatusResettingPendingUserAction,
		models.HostStatusCancelled, // for limited time, until log collection finished or timed-out
		models.HostStatusError,     // for limited time, until log collection finished or timed-out
	}
	for {
		var clusters []*common.Cluster
		timeIt(func() {
			if err = m.db.Preload("Hosts", "status in (?)", monitorStates).Preload(common.MonitoredOperatorsTable).
				Offset(offset).Limit(limit).Order("id").Find(&clusters, "exists (select 1 from hosts where clusters.id = hosts.cluster_id)").Error; err != nil {
				log.WithError(err).Errorf("failed to get clusters")
				return
			}
			//err = m.db.Where("status IN (?)", monitorStates).Offset(offset).Limit(limit).
			//	Order("cluster_id, id").Find(&hosts).Error
		}, "db query")
		if err != nil {
			log.WithError(err).Errorf("failed to get hosts")
			return
		}
		if len(clusters) == 0 {
			break
		}
		for _, c := range clusters {
			for _, host := range c.Hosts {
				if !m.leaderElector.IsLeader() {
					m.log.Debugf("Not a leader, exiting HostMonitoring")
					return
				}
				if !m.SkipMonitoring(host) {
					timeIt(func() {
						err = m.refreshStatusInternal(ctx, host, c, m.db)
					}, "refresh status")
					if err != nil {
						log.WithError(err).Errorf("failed to refresh host %s state", *host.ID)
					}
				}
			}
		}
		offset += limit
	}
}
