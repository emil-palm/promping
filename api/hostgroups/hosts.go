package hostgroups

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mrevilme/promping/api"
	"github.com/mrevilme/promping/config"
	"net/http"
)

func init() {
	api.Router.HandleFunc("/hostgroups/{hostgroup_name}/hosts", ListHostsInHostGroup).Methods("GET")
	api.Router.HandleFunc("/hostgroups/{hostgroup_name}/host", AddHostToHostGroup).Methods("PUT")
	api.Router.HandleFunc("/hostgroups/{hostgroup_name}/host/{name}", DeleteHostFromHostGroup).Methods("DELETE")
	api.Router.HandleFunc("/hostgroups/{hostgroup_name}/hosts", SetHostsInHostGroup).Methods("PATCH")
}

func ListHostsInHostGroup(w http.ResponseWriter, r *http.Request) {
	var hostGroup config.HostGroup
	for _, hostGroup = range config.Current.HostGroups {
		if hostGroup.Name == mux.Vars(r)["hostgroup_name"] {
			json.NewEncoder(w).Encode(hostGroup.Hosts)
			return
		}
	}

	if hostGroup.Name == "" {
		http.Error(w, fmt.Sprintf("Hostgroup '%s' doesn't exist", mux.Vars(r)["hostgroup_name"]), 404)
		return
	}
}

func AddHostToHostGroup(w http.ResponseWriter, r *http.Request) {
	var hostGroup config.HostGroup
	var hostGroupIdx int
	for _hostGroupIdx, _hostGroup := range config.Current.HostGroups {
		if _hostGroup.Name == mux.Vars(r)["hostgroup_name"] {
			hostGroup = _hostGroup
			hostGroupIdx = _hostGroupIdx
		}
	}

	if hostGroup.Name == "" {
		http.Error(w, fmt.Sprintf("Hostgroup '%s' doesn't exist", mux.Vars(r)["hostgroup_name"]), 404)
		return
	}

	decoder := json.NewDecoder(r.Body)
	host := config.Host{}
	err := decoder.Decode(&host)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing post body; %s", err), 500)
		return
	}
	defer r.Body.Close()

	if len(host.Address) <= 0 {
		http.Error(w, "Address is not set", 400)
		return
	}

	if len(host.Name) <= 0 {
		http.Error(w, "Name is not set", 400)
		return
	}

	for _, _host := range hostGroup.Hosts {
		if _host.Name == host.Name {
			http.Error(w, fmt.Sprintf("Host with name '%s' is already present in hostgroup, use PATCH or DELETE", host.Name), 400)
			return
		}
	}

	hostGroup.Hosts = append(hostGroup.Hosts, host)
	config.Current.HostGroups[hostGroupIdx] = hostGroup
	json.NewEncoder(w).Encode(&hostGroup)
	config.Current.Save()
}

func DeleteHostFromHostGroup(w http.ResponseWriter, r *http.Request) {
	var hostGroup config.HostGroup
	var hostGroupIdx int
	for _hostGroupIdx, _hostGroup := range config.Current.HostGroups {
		if _hostGroup.Name == mux.Vars(r)["hostgroup_name"] {
			hostGroup = _hostGroup
			hostGroupIdx = _hostGroupIdx
		}
	}

	if hostGroup.Name == "" {
		http.Error(w, fmt.Sprintf("Hostgroup '%s' doesn't exist", mux.Vars(r)["hostgroup_name"]), 404)
		return
	}

	for hostIdx, _host := range hostGroup.Hosts {
		if _host.Name == mux.Vars(r)["name"] {
			hostGroup.Hosts = append(hostGroup.Hosts[:hostIdx], hostGroup.Hosts[hostIdx+1:]...)
			config.Current.HostGroups[hostGroupIdx] = hostGroup
			config.Current.Save()
			json.NewEncoder(w).Encode(&hostGroup)
			return
		}
	}

	http.Error(w, fmt.Sprintf("Host with name '%s' is not present in hostgroup, use PATCH or PUT", mux.Vars(r)["name"]), 404)
	return
}

func SetHostsInHostGroup(w http.ResponseWriter, r *http.Request) {
	var hostGroup config.HostGroup
	var hostGroupIdx int
	for _hostGroupIdx, _hostGroup := range config.Current.HostGroups {
		if _hostGroup.Name == mux.Vars(r)["hostgroup_name"] {
			hostGroup = _hostGroup
			hostGroupIdx = _hostGroupIdx
		}
	}

	if hostGroup.Name == "" {
		http.Error(w, fmt.Sprintf("Hostgroup '%s' doesn't exist", mux.Vars(r)["hostgroup_name"]), 404)
		return
	}

	decoder := json.NewDecoder(r.Body)
	hosts := make([]config.Host, 0)

	err := decoder.Decode(&hosts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing post body; %s", err), 500)
		return
	}
	defer r.Body.Close()

	hostGroup.Hosts = hosts
	config.Current.HostGroups[hostGroupIdx] = hostGroup
	config.Current.Save()
	json.NewEncoder(w).Encode(&hostGroup)
}
