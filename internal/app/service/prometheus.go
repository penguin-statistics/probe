package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	PromNamespace = "probe"
)

type Prometheus struct {
	pv     *prometheus.CounterVec
	uv     *prometheus.CounterVec
	reconn *prometheus.HistogramVec
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		pv: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: PromNamespace,
			Name:      "page_view_total",
			Help:      "Page views partitioned by platform and path",
		}, []string{"platform", "path"}),
		uv: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: PromNamespace,
			Name:      "unique_view_total",
			Help:      "Unique views partitioned by platform and path, where path is the first page visited",
		}, []string{"platform", "path"}),
		reconn: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: PromNamespace,
			Name:      "reconnection_histogram",
			Help:      "Reconnection values as histogram representing how many times a client has tried to reconnect the service",
			Buckets:   []float64{0, 1, 2, 3, 5, 8, 15, 40, 100, 1000, 10000},
		}, []string{"platform"}),
	}
}

func (p *Prometheus) IncUV(platform, path string) {
	p.uv.WithLabelValues(platform, path).Inc()
}

func (p *Prometheus) IncPV(platform, path string) {
	p.pv.WithLabelValues(platform, path).Inc()
}

func (p *Prometheus) RecordReconnection(platform string, reconnects int) {
	p.reconn.WithLabelValues(platform).Observe(float64(reconnects))
}

//(p *Prometheus)
