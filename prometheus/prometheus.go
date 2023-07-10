package prometheus

import "github.com/prometheus/client_golang/prometheus"

const (
	defaultMetricPath = "/metrics"
)

type MetricType string

const (
	CounterVec   MetricType = "counter_vec"
	Counter      MetricType = "counter"
	GaugeVec     MetricType = "gauge_vec"
	Gauge        MetricType = "gauge"
	HistogramVec MetricType = "histogram_vec"
	Histogram    MetricType = "histogram"
	SummaryVec   MetricType = "summary_vec"
	Summary      MetricType = "summary"
)

type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            MetricType
	Args            []string
}

func NewMetric(m *Metric, serviceName string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case CounterVec:
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case Counter:
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case GaugeVec:
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case Gauge:
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case HistogramVec:
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case Histogram:
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case SummaryVec:
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case Summary:
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: serviceName,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

type Prometheus interface {
}
