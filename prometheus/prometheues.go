package prometheus

import (
	"github.com/mrevilme/promping/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	_ "strings"
	"reflect"
	"strings"
)

type hostGauge struct {
	Host	config.Host
	Gauge 	prometheus.Gauge
}

var hostGauges []hostGauge


func PingGaugeForHost(host config.Host) (prometheus.Gauge) {
	for idx, gauge := range hostGauges {
		if gauge.Host.Name == host.Name {
			if reflect.DeepEqual(gauge.Host.AllTags(), host.AllTags()) {
				return gauge.Gauge
			} else {
				prometheus.Unregister(gauge.Gauge)
				hostGauges = append(hostGauges[:idx], hostGauges[idx+1:]...)
		  	}
		}
	}
	// We didn't find any gauge so we create a new, append it to the cache and return it.
	gauge := hostGauge{
		Host: host,
		Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "promping",
			Name: "rtt",
			ConstLabels: prometheus.Labels{"tags": strings.Join(host.AllTags(),"|"), "host":host.Name},
			Help: "Current route trip time to given device",
		}),
    }
	prometheus.MustRegister(gauge.Gauge)
    hostGauges = append(hostGauges, gauge)
    return gauge.Gauge
}

func Run() {
	http.Handle(config.Current.MetricPath, promhttp.Handler())
}

//var Prom
