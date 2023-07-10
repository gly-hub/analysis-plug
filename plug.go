package analysis

import (
	"github.com/gly-hub/analysis-plug/prometheus"
	routing "github.com/gly-hub/fasthttp-routing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/smallnest/rpcx/server"
	"github.com/spf13/cast"
	"net"
	"net/http"
)

var (
	Config *config
	prom   prometheus.Prometheus
)

type config struct {
	AnalysisServer analysisServer `json:"analysis_server" yaml:"analysisServer"`
}

type analysisServer struct {
	Type        string `json:"type" yaml:"type"`
	Port        int32  `json:"port" yaml:"port"`
	ServiceName string `json:"service_name" yaml:"serviceName"`
	Prometheus  bool   `json:"prometheus" yaml:"prometheus"`
}

func Plug() *Plugin {
	return &Plugin{}
}

type Plugin struct {
}

func (p *Plugin) Config() interface{} {
	Config = &config{}
	return Config
}

func (p *Plugin) InitPlugin() error {
	switch Config.AnalysisServer.Type {
	case "http":
		prom = prometheus.NewHttpPrometheus(Config.AnalysisServer.ServiceName, []*prometheus.Metric{})
	case "rpc":
		prom = prometheus.NewRpcPrometheus(Config.AnalysisServer.ServiceName, []*prometheus.Metric{})
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		listener, _ := net.Listen("tcp", net.JoinHostPort("", cast.ToString(Config.AnalysisServer.Port)))
		_ = http.Serve(listener, nil)
	}()
	return nil
}

func HttpPrometheus() routing.Handler {
	if prom == nil {
		return func(c *routing.Context) error {
			return c.Next()
		}
	}
	switch prom.(type) {
	case *prometheus.HttpPrometheus:
		return prom.(*prometheus.HttpPrometheus).HttpMiddleware()
	}

	return func(c *routing.Context) error {
		return c.Next()
	}
}

func RpcPrometheus() server.Plugin {
	if prom == nil {
		return nil
	}
	switch prom.(type) {
	case *prometheus.RpcPrometheus:
		return prom.(*prometheus.RpcPrometheus).RpcMiddleware()
	}

	return nil
}
