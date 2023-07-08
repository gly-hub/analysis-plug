package analysis

import (
	"github.com/gly-hub/analysis-plug/prometheus"
	routing "github.com/gly-hub/fasthttp-routing"
	"github.com/spf13/cast"
	"net"
	"net/http"
)

var (
	Config *config
	prom   *prometheus.Prometheus
)

type config struct {
	AnalysisServer analysisServer `json:"analysis_server" yaml:"analysisServer"`
}

type analysisServer struct {
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
	prom = prometheus.NewPrometheus(Config.AnalysisServer.ServiceName)
	http.Handle("/metrics", prom.HandlerFunc())
	go func() {
		listener, _ := net.Listen("tcp", net.JoinHostPort("", cast.ToString(Config.AnalysisServer.Port)))
		_ = http.Serve(listener, nil)
	}()
	return nil
}

func PrometheusMiddleware() routing.Handler {
	return prom.HttpMiddleware()
}
