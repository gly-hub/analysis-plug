package prometheus

import (
	routing "github.com/gly-hub/fasthttp-routing"
	"github.com/gly-hub/go-dandelion/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

var (
	httpReqCnt = &Metric{
		ID:          "httpReqCnt",
		Name:        "http_requests_total",
		Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		Type:        CounterVec,
		Args:        []string{"code", "method", "handler", "host", "url"}}

	httpReqDur = &Metric{
		ID:          "httpReqDur",
		Name:        "http_request_duration_seconds",
		Description: "The HTTP request latencies in seconds.",
		Type:        HistogramVec,
		Args:        []string{"code", "method", "url"},
	}

	httpResSz = &Metric{
		ID:          "httpResSz",
		Name:        "http_response_size_bytes",
		Description: "The HTTP response sizes in bytes",
		Type:        Summary,
	}

	httpReqSz = &Metric{
		ID:          "httpReqSz",
		Name:        "http_request_size_bytes",
		Description: "The HTTP request sizes in bytes",
		Type:        Summary,
	}

	httpMetrics = []*Metric{
		httpReqCnt,
		httpReqDur,
		httpResSz,
		httpReqSz,
	}
)

type HttpPrometheus struct {
	httpReqCnt           *prometheus.CounterVec
	httpReqDur           *prometheus.HistogramVec
	httpReqSz, httpResSz prometheus.Summary
	MetricsList          []*Metric
	MetricsPath          string
}

func NewHttpPrometheus(serviceName string, customMetricsList []*Metric) *HttpPrometheus {
	var metricsList []*Metric
	metricsList = append(metricsList, httpMetrics...)
	if len(customMetricsList) > 0 {
		metricsList = append(metricsList, customMetricsList...)
	}

	p := &HttpPrometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
	}

	p.registerMetrics(serviceName)

	return p
}

func (p *HttpPrometheus) registerMetrics(serviceName string) {
	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, serviceName)
		if err := prometheus.Register(metric); err != nil {
			logger.Error("%s could not be registered in Prometheus", metricDef.Name)
		}
		switch metricDef {
		case httpReqCnt:
			p.httpReqCnt = metric.(*prometheus.CounterVec)
		case httpReqDur:
			p.httpReqDur = metric.(*prometheus.HistogramVec)
		case httpResSz:
			p.httpResSz = metric.(prometheus.Summary)
		case httpReqSz:
			p.httpReqSz = metric.(prometheus.Summary)
		}
		metricDef.MetricCollector = metric
	}
}

func (p *HttpPrometheus) HttpMiddleware() routing.Handler {
	return func(c *routing.Context) error {
		if string(c.Path()) == p.MetricsPath {
			return c.Next()
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.RequestCtx)

		_ = c.Next()

		status := strconv.Itoa(c.Response.StatusCode())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(len(c.Response.Body()))

		url := string(c.Path())
		p.httpReqDur.WithLabelValues(status, string(c.Request.Header.Method()), url).Observe(elapsed)
		p.httpReqCnt.WithLabelValues(status, string(c.Request.Header.Method()), c.Request.String(), string(c.Request.Header.Host()), url).Inc()
		p.httpReqSz.Observe(float64(reqSz))
		p.httpResSz.Observe(resSz)
		return nil
	}
}

func computeApproximateRequestSize(r *fasthttp.RequestCtx) int {
	s := 0
	if r.URI() != nil {
		s = len(string(r.URI().Path()))
	}

	s += len(string(r.Request.Header.Method()))
	s += len(string(r.Request.Header.Protocol()))

	headers := make(map[string]string)
	r.Request.Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	for name, values := range headers {
		s += len(name)
		s += len(values)
	}
	s += len(string(r.Host()))

	if r.Request.Header.ContentLength() != -1 {
		s += r.Request.Header.ContentLength()
	}
	return s
}
