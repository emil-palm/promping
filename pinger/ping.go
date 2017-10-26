package pinger

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/bastjan/go-ping"
	"github.com/mrevilme/promping/config"
	"github.com/mrevilme/promping/prometheus"
	"net"
	"time"
)

func worker(id int, jobs chan config.Host) {
	for host := range jobs {
		log.Infof("worker %d started  job", id)
		pingHost(host)
		log.Infof("worker %d finished job", id)
	}
}

func Run() {
	jobs := make(chan config.Host, 100)
	go func(_config *config.Config, jobs chan config.Host) {
		for {
			executePings(jobs)
			time.Sleep(time.Second * 10)
		}
	}(&config.Current,jobs)
	for w := 1; w <= config.Current.Workers; w++ {
		log.Debugf("Starting worker %d", w)
		go worker(w, jobs)
	}
}

func executePings(jobs chan config.Host) {
	for _, hostgroup := range config.Current.HostGroups {
		for _, host := range hostgroup.Hosts {
			host.SetHostGroup(hostgroup)
			//go pingHost(host)
			if host.ShouldUpdate() {
				jobs <- host
				host.SetLostPoll(time.Now())
			}
		}
	}
}

func pingHost(host config.Host) {
	log.Debugf("Pinging %s", host.Name)

	var network string
	netAddress := net.ParseIP(host.Address)
	if netAddress != nil {

		if netAddress.To4() != nil {
			network = "ip4:icmp"
		} else {
			network = "ip6:icmp"
		}
	} else {
		if proto, _ := config.ParseProtocol(host.Protocol); proto == config.IPv6 {
			network = "ip6:icmp"
		} else {
			network = "ip4:icmp"
		}
	}

	// Execute ping
	ctx := context.Background()
	pinger, err := ping.NewPingerWithNetwork(ctx, host.Address, network)
	if err != nil {
		log.Error(err)
		return
	}

	pinger.Timeout = time.Second * 1
	pinger.Count = 3
	pinger.SetPrivileged(true)
	pinger.Run()

	stats := pinger.Statistics()
	gauge := prometheus.PingGaugeForHost(host)

	gauge.PacketLossGauage.Set(stats.PacketLoss)
	gauge.MinRTTGauge.Set(stats.MinRtt.Seconds())
	gauge.MaxRTTGauge.Set(stats.MaxRtt.Seconds())
	gauge.AvgRTTGauge.Set(stats.AvgRtt.Seconds())
	gauge.StdDevRTTGauge.Set(stats.StdDevRtt.Seconds())
	//gauge.Set(float64(stats.Seconds()))
}
