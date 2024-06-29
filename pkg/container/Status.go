package container

import (
	"github.com/qdnqn/smr/pkg/static"
	"time"
)

func (container *Container) UpdateStatus(status int, value bool) {
	switch status {
	case static.STATUS_CREATED:
		container.Status.Created = value
		container.Status.CreatedTime = time.Now()
	case static.STATUS_READINESS:
		container.Status.Readiness = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_READINESS_FAILED:
		container.Status.ReadinessFailed = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_READY:
		container.Status.Ready = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_DEPENDS_SOLVED:
		container.Status.DependsSolved = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_HEALTHY:
		container.Status.Healthy = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_RUNNING:
		container.Status.Running = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_RECONCILING:
		container.Status.Reconciling = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_DRIFTED:
		container.Status.DefinitionDrift = value
		container.Status.LastUpdate = time.Now()
		break
	case static.STATUS_PENDING_DELETE:
		container.Status.PendingDelete = value
		container.Status.LastUpdate = time.Now()
		break
	default:
		break
	}
}
