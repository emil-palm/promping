package config

import (
	"strings"
	"fmt"
)

var Channel chan Config

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
}

func (h *Host) AllTags(hg HostGroup) []string {
	return append(h.Tags, hg.Tags...)
}

func init() {
	Channel = make(chan Config)
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