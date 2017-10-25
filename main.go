package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/mrevilme/promping/config"
	"github.com/mrevilme/promping/pinger"
	_ "github.com/mrevilme/promping/prometheus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	/*
	  "github.com/jaxxstorm/graphping/config"
	  "github.com/jaxxstorm/graphping/ping"
	  "github.com/prometheus/client_golang/prometheus/promhttp"
	  "net/http"
	*/)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr, could also be a file.
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Configure flag support

	pflag.String("config", "promping", "configuration file to use")
	pflag.String("httplisten", ":8080", "Address to expose http server on")
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

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

	log.Debug("Starting pinger")
	pinger.Run()

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
