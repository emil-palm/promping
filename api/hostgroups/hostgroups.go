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
