package modules

import (
	"github.com/prometheus/prometheus/model/labels"
	"github.com/roidelapluie/o11y-deploy/model/promserver"
)

type PrometheusModule interface {
	GetPrometheusServers([]labels.Labels, string) ([]promserver.PrometheusServer, error)
}
