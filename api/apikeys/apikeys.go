package apikeys

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
	api.Router.HandleFunc("/keys", ListKeys).Methods("GET")
	api.Router.HandleFunc("/keys", AddKey).Methods("PUT")
	api.Router.HandleFunc("/keys/{key}", DeleteKey).Methods("DELETE")
}

func DeleteKey(w http.ResponseWriter, r *http.Request) {
	for idx, apikey := range config.Current.ApiKeys {
		if apikey == mux.Vars(r)["key"] {
			log.Debugf("Removing %s at index: %d", apikey, idx)
		}
	}

	http.Error(w, fmt.Sprintf("Key '%s' doesn't exist", mux.Vars(r)["key"]), 404)
	return
}

func AddKey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	key := make(map[string]string)
	err := decoder.Decode(&key)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	defer r.Body.Close()

	for _, apikey := range config.Current.ApiKeys {
		if apikey == key["key"] {
			http.Error(w, fmt.Sprintf("Hostgroup with name '%s' is already present, use PATCH or DELETE", apikey), 400)
			return
		}
	}
	config.Current.ApiKeys = append(config.Current.ApiKeys, key["key"])

	json.NewEncoder(w).Encode(&config.Current.ApiKeys)
	config.Current.Save()
}

func ListKeys(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(config.Current.ApiKeys)
}
