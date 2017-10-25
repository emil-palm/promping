package pinger

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mrevilme/promping/config"
	"github.com/mrevilme/promping/prometheus"
	"net"
	"time"
	"github.com/tatsushid/go-fastping"
)

type response struct {
	host config.Host
	addr *net.IPAddr
	rtt  time.Duration
}

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
	// Execute ping
	pinger := fastping.NewPinger()

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
	ra, err := net.ResolveIPAddr(network, host.Address)
	if err != nil {
		log.Error(err)
		return
	}

	pinger.AddIPAddr(ra)
	maxTTLExceeded := true
	var rtt time.Duration
	pinger.OnRecv = func(addr *net.IPAddr, _rtt time.Duration) {
		rtt = _rtt
		maxTTLExceeded = false
		log.Debugf("IP Addr: %s receive, RTT: %v", addr.String(), _rtt)
	}
	pinger.OnIdle = func() {
		if maxTTLExceeded == false {
			log.Infof("Finished %s with RTT: %v", host.Name, rtt)
			gauge := prometheus.PingGaugeForHost(host)
			gauge.Set(float64(rtt.Seconds()));
		} else {
			log.Infof("Timeout %s", host.Name)
		}
	}
	err = pinger.Run()
	if err != nil {
		log.Error(err)
	}
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
