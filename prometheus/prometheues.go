package prometheus

import (
	"github.com/mrevilme/promping/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"reflect"
	"strings"
	_ "strings"
)

type HostGauge struct {
	Host             config.Host
	PacketLossGauage prometheus.Gauge
	MinRTTGauge      prometheus.Gauge
	MaxRTTGauge      prometheus.Gauge
	AvgRTTGauge      prometheus.Gauge
	StdDevRTTGauge   prometheus.Gauge
}

var hostGauges []HostGauge

func PingGaugeForHost(host config.Host) *HostGauge {
	for idx, gauge := range hostGauges {
		if gauge.Host.Name == host.Name {
			if reflect.DeepEqual(gauge.Host.AllTags(), host.AllTags()) {
				return &gauge
			} else {
				prometheus.Unregister(gauge.PacketLossGauage)
				prometheus.Unregister(gauge.MinRTTGauge)
				prometheus.Unregister(gauge.MaxRTTGauge)
				prometheus.Unregister(gauge.AvgRTTGauge)
				prometheus.Unregister(gauge.StdDevRTTGauge)
				hostGauges = append(hostGauges[:idx], hostGauges[idx+1:]...)
			}
		}
	}
	// We didn't find any gauge so we create a new, append it to the cache and return it.
	gauge := HostGauge{
		Host: host,
		PacketLossGauage: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "promping",
			Name:        "packetloss",
			ConstLabels: prometheus.Labels{"tags": strings.Join(host.AllTags(), "|"), "host": host.Name},
			Help:        "Packetloss percentage",
		}),
		MinRTTGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "promping",
			Name:        "rtt",
			Subsystem:   "min",
			ConstLabels: prometheus.Labels{"tags": strings.Join(host.AllTags(), "|"), "host": host.Name},
			Help:        "Minimum route-trip time",
		}),
		MaxRTTGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "promping",
			Name:        "rtt",
			Subsystem:   "max",
			ConstLabels: prometheus.Labels{"tags": strings.Join(host.AllTags(), "|"), "host": host.Name},
			Help:        "Maximum route-trip time",
		}),
		AvgRTTGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "promping",
			Name:        "rtt",
			Subsystem:   "avg",
			ConstLabels: prometheus.Labels{"tags": strings.Join(host.AllTags(), "|"), "host": host.Name},
			Help:        "Avergage route-trip time",
		}),
		StdDevRTTGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "promping",
			Subsystem:   "stdev",
			Name:        "rtt",
			ConstLabels: prometheus.Labels{"tags": strings.Join(host.AllTags(), "|"), "host": host.Name},
			Help:        "Standard deviation of the RTT",
		}),
	}
	prometheus.MustRegister(gauge.PacketLossGauage)
	prometheus.MustRegister(gauge.MinRTTGauge)
	prometheus.MustRegister(gauge.MaxRTTGauge)
	prometheus.MustRegister(gauge.AvgRTTGauge)
	prometheus.MustRegister(gauge.StdDevRTTGauge)
	hostGauges = append(hostGauges, gauge)
	return &gauge
}

func Run() {
	http.Handle(config.Current.MetricPath, promhttp.Handler())
}
