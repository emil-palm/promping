package config

import (
	"strings"
	"fmt"
)

var Channel chan Config
var Current Config

type Config struct {
	HostGroups 	[]HostGroup
	HTTPListen 	string
	MetricPath 	string
	Loglevel 	string
}

type HostGroup struct {
	Name  string
	Hosts []Host
	Tags  []string
}

type Host struct {
	Address string
	Name    string
	Protocol string
	Tags    []string
	HostGroup *HostGroup
}

func (h *Host) AllTags() []string {
	return append(append(h.Tags, h.HostGroup.Tags...),h.HostGroup.Name)
}

func init() {
	Channel = make(chan Config)
	go func(cfg *Config) {
		for {
			*cfg = <-Channel
		}
	}(&Current)
}

type Protocol uint32

func ParseProtocol(protocol string) (Protocol, error) {

	switch strings.ToLower(protocol) {
		case "ipv4":
			return IPv4, nil
		case "ipv6":
			return IPv6,nil
	}
	var l Protocol
	return l, fmt.Errorf("not a valid protocol: %q", protocol)
}

const (
	IPv4 Protocol = iota
	IPv6
)