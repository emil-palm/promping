package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/mrevilme/promping/api"
	_ "github.com/mrevilme/promping/api/hostgroups"
	"github.com/mrevilme/promping/config"
	"github.com/mrevilme/promping/pinger"
	"github.com/mrevilme/promping/prometheus"
	_ "github.com/mrevilme/promping/prometheus"
	"github.com/spf13/pflag"
	"github.com/theherk/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr, could also be a file.
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.

	// Configure flag support
	pflag.String("config", "promping", "configuration file to use")
	pflag.String("httplisten", ":8080", "Address to expose http server on")
	pflag.String("loglevel", "debug", "Loglevel to use")
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

	// Configure viper options
	viper.SetDefault("metricpath", "/ping")
	viper.SetDefault("httplisten", ":8080")
	viper.SetDefault("loglevel", "warn")

	// Setup viper for configuration handling
	// optionally look for config in the working directory
	if _config, err := pflag.CommandLine.GetString("config"); err == nil {
		viper.SetConfigFile(_config)
	} else {
		viper.SetConfigName(viper.GetString("config")) // name of config file (without extension)
		viper.AddConfigPath("/etc/promping/")          // path to look for the config file in
		viper.AddConfigPath("$HOME/.promping")         // call multiple times to add many search paths
		viper.AddConfigPath(".")
	}

	if loglevel, err := pflag.CommandLine.GetString("loglevel"); err == nil {
		level, err := log.ParseLevel(loglevel)
		if err != nil {
			log.Panic(err)
		}
		log.SetLevel(level)
	}
}

func main() {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	err := viper.ReadInConfig() // Find and read the config file

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	readConfig()

	// we watch for configuration changes, if push the new config into a channel to be handled by gorutines
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debugf("Config file changed:", e.Name)
		readConfig()
	})

	for {
		if config.Current.MetricPath != "" {
			break
		}
	}

	log.Debug("Starting pinger")
	pinger.Run()
	prometheus.Run()
	api.Run()
	http.ListenAndServe(config.Current.HTTPListen, nil)

	// Waiting for interupt
	c := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	go func() {
		for sig := range c {
			log.Warn(fmt.Sprintf("captured %v", sig))
			done <- true
		}
	}()
	<-done
}

func readConfig() {
	_config := config.Config{}
	err := viper.Unmarshal(&_config)
	if err != nil {
		log.Panic(err)
	}
	go func() {
		config.Channel <- _config
	}()
}
