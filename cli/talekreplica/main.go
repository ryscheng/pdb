package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/etcd/pkg/flags"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server"
	"github.com/spf13/pflag"
)

// Starts a single, centralized talek replica operating with configuration from talekutil
func main() {
	log.Println("---------------------")
	log.Println("--- Talek Replica ---")
	log.Println("---------------------")

	// Support setting flags from either command-line arguments or environment variables
	// command-line arguments take priority
	configPath := pflag.StringP("config", "c", "replica.conf", "Talek Replica Configuration (env TALEK_CONFIG)")
	commonPath := pflag.StringP("common", "f", "common.conf", "Talek Common Configuration (env TALEK_COMMON)")
	backing := pflag.StringP("backing", "b", "cpu.0", "PIR daemon method (env TALEK_BACKING)")
	listen := pflag.StringP("listen", "l", ":8080", "Listening Address")
	err := flags.SetPflagsFromEnv(common.EnvPrefix, pflag.CommandLine)
	if err != nil {
		log.Printf("Error reading environment variables, %v\n", err)
		return
	}
	pflag.Parse()

	log.Printf("Arguments:\n")
	log.Printf("config=%v\n", *configPath)
	log.Printf("backing=%v\n", *backing)

	configString, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Printf("Could not read %s!\n", *configPath)
		return
	}
	commonString, err := ioutil.ReadFile(*commonPath)
	if err != nil {
		log.Printf("Could not read %s!\n", *commonPath)
		return
	}

	// Default configuration. The server can be started with just a trustdomain
	// config and this will be used for the serverConfig struct in that case.
	serverConfig := server.Config{
		Config:           &common.Config{},
		WriteInterval:    time.Second,
		ReadInterval:     time.Second,
		ReadBatch:        8,
		TrustDomain:      &common.TrustDomainConfig{},
		TrustDomainIndex: 0,
	}
	if err = json.Unmarshal(configString, &serverConfig); err != nil {
		log.Printf("Could not parse %s: %v\n", *configPath, err)
		return
	}
	if err = json.Unmarshal(commonString, serverConfig.Config); err != nil {
		log.Printf("Could not parse %s: %v\n", *commonPath, err)
		return
	}

	log.Printf("Using the following configuration:")
	log.Printf("serverConfig=%#+v\n", serverConfig)
	log.Printf("serverConfig.Config=%#+v\n", serverConfig.Config)

	r := server.NewReplicaServer(serverConfig.TrustDomain.Name, *backing, serverConfig)
	listener, err := r.Run(*listen)
	if err != nil {
		log.Printf("Couldn't listen to frontend address: %v\n", err)
		return
	}

	log.Println("Running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	r.Replica.Close()
	listener.Close()
}
