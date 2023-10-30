package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/team-dandelion/go-dandelion/logger"
)

var (
	rpcReqCnt = &Metric{
		ID:          "rpcReqCnt",
		Name:        "rpc_requests_total",
		Description: "How many RPC requests processed",
		Type:        CounterVec,
	}

	rpcResSz = &Metric{
		ID:          "rpcResSz",
		Name:        "rpc_response_size_bytes",
		Description: "The RPC response sizes in bytes",
		Type:        Summary,
	}

	rpcReqSz = &Metric{
		ID:          "rpcReqSz",
		Name:        "rpc_request_size_bytes",
		Description: "The RPC request sizes in bytes",
		Type:        Summary,
	}

	rpcMetrics = []*Metric{
		rpcReqCnt,
		rpcResSz,
		rpcReqSz,
	}
)

type RpcPrometheus struct {
	rpcReqCnt          *prometheus.CounterVec
	rpcReqSz, rpcResSz prometheus.Summary
	MetricsList        []*Metric
	MetricsPath        string
}

func NewRpcPrometheus(serviceName string, customMetricsList []*Metric) *RpcPrometheus {
	var metricsList []*Metric
	metricsList = append(metricsList, rpcMetrics...)
	if len(customMetricsList) > 0 {
		metricsList = append(metricsList, customMetricsList...)
	}

	p := &RpcPrometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
	}

	p.registerMetrics(serviceName)

	return p
}

func (p *RpcPrometheus) registerMetrics(serviceName string) {
	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, serviceName)
		if err := prometheus.Register(metric); err != nil {
			logger.Error("%s could not be registered in Prometheus", metricDef.Name)
		}
		switch metricDef {
		case rpcReqCnt:
			p.rpcReqCnt = metric.(*prometheus.CounterVec)
		case rpcReqSz:
			p.rpcReqSz = metric.(prometheus.Summary)
		case rpcResSz:
			p.rpcResSz = metric.(prometheus.Summary)
		}
		metricDef.MetricCollector = metric
	}
}

func (p *RpcPrometheus) RpcMiddleware() server.Plugin {
	return &RpcPrometheusPlugin{
		rpcReqCnt: p.rpcReqCnt,
		rpcReqSz:  p.rpcReqSz,
		rpcResSz:  p.rpcResSz,
	}
}

type RpcPrometheusPlugin struct {
	rpcReqCnt          *prometheus.CounterVec
	rpcReqSz, rpcResSz prometheus.Summary
}

func (rpc *RpcPrometheusPlugin) PreHandleRequest(ctx context.Context, r *protocol.Message) error {
	return nil
}

func (rpc *RpcPrometheusPlugin) PostWriteResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	rpc.rpcReqSz.Observe(float64(len(req.Encode())))
	rpc.rpcResSz.Observe(float64(len(res.Encode())))
	rpc.rpcReqCnt.WithLabelValues().Inc()
	return nil
}
