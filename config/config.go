package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"github.com/imdario/mergo"
	"github.com/theherk/viper"
	"strings"
	"time"
)

var Channel chan Config
var Current Config

type Config struct {
	HostGroups []HostGroup
	HTTPListen string
	MetricPath string
	Loglevel   string
	ApiKeys    []string
	Interval   int
	Workers    int
}

func (c *Config) Save() {
	_map := structs.Map(c)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(_map)
	err := viper.MergeConfig(&buf)
	if err != nil {
		log.Panic(err)
	}

	viper.WriteConfig()
}

type HostGroup struct {
	Name  string
	Hosts []Host
	Tags  []string `json:",omitempty" mapstructure:","`
	Interval int
}

func (hg *HostGroup) Merge(hostGroup HostGroup) error {
	return mergo.Merge(hg, hostGroup)
}

type Host struct {
	Address   string
	Name      string
	Protocol  string    `json:",omitempty" mapstructure:","`
	Tags      []string  `json:",omitempty" mapstructure:","`
	Interval  int
	hostGroup HostGroup `json:"-"`
	lastPoll  time.Time `json:",omitempty" mapstructure:","`
}

func (h *Host) AllTags() []string {
	return append(append(h.Tags, h.hostGroup.Tags...), h.hostGroup.Name)
}

func (h *Host) SetHostGroup(hostGroup HostGroup) {
	h.hostGroup = hostGroup
}

func (h *Host) SetLostPoll(lastUpdate time.Time) {
	h.lastPoll = lastUpdate
}

func (h *Host) ShouldUpdate() bool {
	if h.lastPoll.IsZero() {
		return true
	}
	var interval time.Duration = time.Duration(Current.Interval)*time.Second

	if time.Duration(h.Interval) > 0 {
		interval = time.Duration(h.Interval)*time.Second
	} else if time.Duration(h.hostGroup.Interval) > interval {
		interval = time.Duration(h.hostGroup.Interval)*time.Second
	}

	log.Warnf("%s", h.lastPoll)
	return time.Now().Sub(h.lastPoll) > interval
}

func (h *Host) Merge(host Host) error {
	return mergo.Merge(h, host)
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
		return IPv6, nil
	}
	var l Protocol
	return l, fmt.Errorf("not a valid protocol: %q", protocol)
}

const (
	IPv4 Protocol = iota
	IPv6
)
