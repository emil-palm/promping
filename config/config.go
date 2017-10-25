package config

var Channel chan Config

type Config struct {
	HostGroups []HostGroup
	HTTPListen string
}

type HostGroup struct {
	Name  string
	Hosts []Host
	Tags  []string
}

type Host struct {
	Address string
	Name    string
	Tags    []string
}

func (h *Host) AllTags(hg HostGroup) []string {
	return append(h.Tags, hg.Tags...)
}

func init() {
	Channel = make(chan Config)
}
