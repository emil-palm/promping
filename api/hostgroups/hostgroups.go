package hostgroups

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mrevilme/promping/api"
	"github.com/mrevilme/promping/config"
	"net/http"
)

func init() {
	api.Router.HandleFunc("/hostgroups", GetHostGroups).Methods("GET")
	api.Router.HandleFunc("/hostgroups", AddHostGroup).Methods("PUT")
	api.Router.HandleFunc("/hostgroups/{name}", DeleteHostGroup).Methods("DELETE")
	api.Router.HandleFunc("/hostgroups/{hostgroup_name}/host", AddHostToHostGroup).Methods("PUT")
	api.Router.HandleFunc("/hostgroups/{hostgroup_name}", PatchHostGroup).Methods("PATCH")
}

func DeleteHostGroup(w http.ResponseWriter, r *http.Request) {
	for idx, hostGroup := range config.Current.HostGroups {
		if hostGroup.Name == mux.Vars(r)["name"] {
			log.Debugf("Removing %s at index: %d", hostGroup.Name, idx)
		}
	}
}

func AddHostGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	hostGroup := config.HostGroup{}
	err := decoder.Decode(&hostGroup)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	defer r.Body.Close()

	for _, _hostGroup := range config.Current.HostGroups {
		if hostGroup.Name == _hostGroup.Name {
			http.Error(w, fmt.Sprintf("Hostgroup with name '%s' is already present, use PATCH or DELETE", hostGroup.Name), 400)
			return
		}
	}
	config.Current.HostGroups = append(config.Current.HostGroups, hostGroup)

	json.NewEncoder(w).Encode(&hostGroup)
	config.Current.Save()
}

func GetHostGroups(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(config.Current.HostGroups)
}

func PatchHostGroup(w http.ResponseWriter, r *http.Request) {
	var _hostGroup config.HostGroup
	var hostGroupIdx int
	for _hostGroupIdx, hostGroup := range config.Current.HostGroups {
		if hostGroup.Name == mux.Vars(r)["hostgroup_name"] {
			_hostGroup = hostGroup
			hostGroupIdx = _hostGroupIdx
		}
	}

	if _hostGroup.Name == "" {
		http.Error(w, fmt.Sprintf("Hostgroup '%s' doesn't exist", mux.Vars(r)["hostgroup_name"]), 400)
		return
	}

	decoder := json.NewDecoder(r.Body)
	hostGroup := config.HostGroup{}

	err := decoder.Decode(&hostGroup)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing post body; %s", err), 500)
		return
	}
	defer r.Body.Close()

	err = hostGroup.Merge(_hostGroup)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error merging objects; %s", err), 500)
		return
	}

	config.Current.HostGroups[hostGroupIdx] = hostGroup
	config.Current.Save()
	json.NewEncoder(w).Encode(&hostGroup)
}
