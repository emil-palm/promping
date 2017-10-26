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

func Run() {
	go func(_config *config.Config) {
		for {
			executePings()
			time.Sleep(time.Second * 10)
		}
	}(&config.Current)
}

func executePings() {
	for _, hostgroup := range config.Current.HostGroups {
		for _, host := range hostgroup.Hosts {
			host.HostGroup = &hostgroup
			go pingHost(host)
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

/*
func RunPinger(interval int) error {

	// create a new pinger
	p :=

	// create some interfaces to store results
	results := make(map[string]*response)
	index := make(map[string]string)

	// loop througb the targets in the struct and resolve the address
	for _, target := range group.Targets {
		ra, err := net.ResolveIPAddr("ip4:icmp", target.Address)
		log.Debug(fmt.Sprintf("Target %s resolves to %s", target.Address, ra))
		if err != nil {
			return errors.New(fmt.Sprintf("Can't resolve %s", target.Address))
		}

		// store the result of each ping poll
		results[ra.String()] = nil

		// map the ip address back to the label
		index[ra.String()] = target.Label

		// add the IP address to the list of ping endpoints
		p.AddIPAddr(ra)
	}

	onRecv, onIdle := make(chan *response), make(chan bool)
	p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{addr: addr, rtt: t}
	}
	p.OnIdle = func() {
		onIdle <- true
	}

	//determine what interval we should run at
	if group.Interval > 0 {
		log.Debug(fmt.Sprintf("Group Interval defined: %v", group.Interval))
		p.MaxRTT = time.Duration(group.Interval) * time.Second
	} else if interval > 0 {
		log.Debug(fmt.Sprintf("Global Interval defined: %v", interval))
		p.MaxRTT = time.Duration(interval) * time.Second
	} else {
		log.Debug(fmt.Sprintf("Using Default Interval: 60"))
		p.MaxRTT = 60 * time.Second
	}

	// set the metric path
	metricPath := fmt.Sprintf("%s", group.Prefix)
	log.Info("Starting pinger for ", group.Name, " with metric path: ", metricPath)
	p.RunLoop()

pingloop:
	for {
		select {
		case res := <-onRecv:
			if _, ok := results[res.addr.String()]; ok {
				results[res.addr.String()] = res
			}
		case <-onIdle:
			for host, r := range results {
				outputLabel := index[host]
				if r == nil {
					log.Debug(fmt.Sprintf("%s : unreachable", outputLabel))
					// send a metric for a failed ping
					err := statsdClient.Inc(fmt.Sprintf("%s.failed", outputLabel), 1, 1)
					if err != nil {
						log.Error(fmt.Sprintf("Error sending metric: %+v", err))
					}
				} else {
					log.Debug(fmt.Sprintf("%s : %v", outputLabel, r.rtt))
					// send a zeroed failed metric, because we succeeded!
					err := statsdClient.Inc(fmt.Sprintf("%s.failed", outputLabel), 0, 1)
					if err != nil {
						log.Error(fmt.Sprintf("Error sending metric: %+v", err))
					}
					err = statsdClient.TimingDuration(fmt.Sprintf("%s.timer", outputLabel), r.rtt, 1)
				}
				results[host] = nil
			}
		case <-p.Done():
			if err := p.Err(); err != nil {
				return errors.New("Can't start pinger")
			}
			break pingloop
		}
	}
	p.Stop()

	return errors.New("failed")

}*/
