package connectivity

import (
	"encoding/json"
	"time"

	"github.com/openshift/assisted-service/models"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=validator.go -package=connectivity -destination=mock_connectivity_validator.go
type Validator interface {
	GetHostValidInterfaces(host *models.Host) ([]*models.Interface, error)
}

func NewValidator(log logrus.FieldLogger) Validator {
	return &validator{
		log:   log,
		cache: cache.New(time.Hour, time.Hour),
	}
}

type validator struct {
	log   logrus.FieldLogger
	cache *cache.Cache
}

func getKey(host *models.Host) string {
	return host.ID.String() + "@" + host.InfraEnvID.String()
}

func (v *validator) GetHostValidInterfaces(host *models.Host) ([]*models.Interface, error) {
	key := getKey(host)
	val, exists := v.cache.Get(key)
	if exists {
		return val.([]*models.Interface), nil
	}
	var inventory models.Inventory
	if err := json.Unmarshal([]byte(host.Inventory), &inventory); err != nil {
		return nil, err
	}
	if len(inventory.Interfaces) == 0 {
		return nil, errors.Errorf("host %s doesn't have interfaces", host.ID)
	}
	v.cache.Set(key, inventory.Interfaces, cache.DefaultExpiration)
	return inventory.Interfaces, nil
}
