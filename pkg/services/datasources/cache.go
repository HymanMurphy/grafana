package datasources

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/cache"
)

type CacheService interface {
	GetDatasource(datasourceID int64, user *m.SignedInUser, skipCache bool) (*m.DataSource, error)
}

type CacheServiceImpl struct {
	Bus          bus.Bus             `inject:""`
	CacheService *cache.CacheService `inject:""`
}

func init() {
	registry.RegisterService(&CacheServiceImpl{})
}

func (dc *CacheServiceImpl) Init() error {
	return nil
}

func (dc *CacheServiceImpl) GetDatasource(datasourceID int64, user *m.SignedInUser, skipCache bool) (*m.DataSource, error) {
	cacheKey := fmt.Sprintf("ds-%d", datasourceID)

	if !skipCache {
		if cached, found := dc.CacheService.Get(cacheKey); found {
			ds := cached.(*m.DataSource)
			if ds.OrgId == user.OrgId {
				return ds, nil
			}
		}
	}

	query := m.GetDataSourceByIdQuery{Id: datasourceID, OrgId: user.OrgId}
	if err := dc.Bus.Dispatch(&query); err != nil {
		return nil, err
	}

	dc.CacheService.Set(cacheKey, query.Result, time.Second*5)
	return query.Result, nil
}
