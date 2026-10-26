package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	PermissionCheckTotal    *prometheus.CounterVec
	PermissionCheckDuration *prometheus.HistogramVec
	HTTPRequestTotal        *prometheus.CounterVec
	HTTPRequestDuration     *prometheus.HistogramVec
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		PermissionCheckTotal:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "contract_permission_check_total", Help: "Permission checks."}, []string{"result"}),
		PermissionCheckDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "contract_permission_check_duration_seconds", Help: "Permission check duration."}, []string{"result"}),
		HTTPRequestTotal:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "contract_http_requests_total", Help: "HTTP requests."}, []string{"method", "route", "status"}),
		HTTPRequestDuration:     prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "contract_http_request_duration_seconds", Help: "HTTP request duration."}, []string{"method", "route", "status"}),
	}
	reg.MustRegister(m.PermissionCheckTotal, m.PermissionCheckDuration, m.HTTPRequestTotal, m.HTTPRequestDuration)
	for _, result := range []string{"allowed", "denied", "error"} {
		m.PermissionCheckTotal.WithLabelValues(result)
		m.PermissionCheckDuration.WithLabelValues(result)
	}
	return m
}
